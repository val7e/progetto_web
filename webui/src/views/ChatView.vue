<script>
export default {
    data: function() {
        return {
            errormsg: null,
            loading: false,
            conversation: null,
            messages: [],
            newMessage: '',
            selectedPhoto: null,
            photoPreview: null,
            newComment: {},
            showComments: {},
            messageComments: {},
            currentUsername: localStorage.getItem('username'),
            recipientPhoto: null,
            // Group settings state
            showGroupSettings: false,
            editingGroupName: false,
            newGroupName: '',
            groupPhotoPreview: null,
            groupSelectedPhoto: null,
            newGroupMember: '',
            groupMemberSearchQuery: '',
            groupMemberSearchResults: [],
            groupMemberSearchLoading: false,
            contextMenu: {
                show: false,
                x: 0,
                y: 0,
                messageId: null,
                isOwnMessage: false
            },
            showForwardDialog: false,
            forwardMessageId: null,
            availableConversations: [],
            selectedConversationId: null,
            // forward search
            forwardSearchQuery: '',
            forwardSearchLoading: false,
            forwardSearchResults: []
        }
    },
    methods: {
        async refresh() {
            this.loading = true;
            this.errormsg = null;
            try {
                const conversationId = this.$route.params.conversationId;
                let response = await this.$axios.get(`/conversations/${conversationId}`);
                this.conversation = response.data;
                this.messages = (response.data.messages || []).slice().sort((a,b) => {
                    const ta = a.timestamp ? new Date(a.timestamp).getTime() : 0;
                    const tb = b.timestamp ? new Date(b.timestamp).getTime() : 0;
                    return tb - ta;
                });
                if (this.conversation?.type === 'group') {
                    this.newGroupName = this.conversation?.name || ''
                    const convId = this.$route.params.conversationId
                    try {
                        const g = await this.$axios.get(`/groups/${convId}`)
                        if (g.data?.name) this.conversation.name = g.data.name
                        if (g.data?.group_photo) this.conversation.convo_pic = g.data.group_photo
                        //  update participants
                        if (g.data?.members) this.conversation.participants = g.data.members
                    } catch (e) {
                    }
                } else {
                    await this.fetchRecipientPhoto();
                }
            } catch (e) {
                this.errormsg = e.toString();
            }
            this.loading = false;
        }
,
        async fetchRecipientPhoto() {
            const recipient = this.getOtherParticipant();
            if (!recipient) return;
            
            try {
                let response = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(recipient)}`);
                if (response.data && response.data.length > 0) {
                    const user = response.data.find(u => u.username === recipient);
                    if (user && user.pic) {
                        this.recipientPhoto = user.pic;
                    }
                }
            } catch (e) {
                console.error('Failed to fetch recipient photo:', e);
            }
        },
        getOtherParticipant() {
            const others = this.conversation?.participants?.filter(p => p !== this.currentUsername);
            return others?.join(', ') || 'Unknown';
        },
                
        handleFileSelect(event) {
            const file = event.target.files[0];
            if (!file) return;
            
            if (!file.type.startsWith('image/')) {
                this.errormsg = "Please select an image file";
                return;
            }
            
            this.selectedPhoto = file;
            
            const reader = new FileReader();
            reader.onload = (e) => {
                this.photoPreview = e.target.result;
            };
            reader.readAsDataURL(file);
        },
        
        clearPhoto() {
            this.selectedPhoto = null;
            this.photoPreview = null;
            this.$refs.photoInput.value = '';
        },
        
        async sendMessage() {
            if (this.selectedPhoto) {
                await this.sendPhoto();
                return;
            }
            
            if (!this.newMessage.trim()) return;
            
            this.errormsg = null;
            try {
                const conversationId = this.$route.params.conversationId;
                await this.$axios.post(`/conversations/${conversationId}/messages`, {
                    type: "text",
                    text: this.newMessage
                });
                this.newMessage = '';
                await this.refresh();
            } catch (e) {
                this.errormsg = e.toString();
            }
        },
        
        async sendPhoto() {
            if (!this.selectedPhoto) return;
            
            this.errormsg = null;
            this.loading = true;
            
            try {
                const reader = new FileReader();
                reader.onload = async (e) => {
                    const base64Photo = e.target.result.split(',')[1];
                    
                    const conversationId = this.$route.params.conversationId;
                    await this.$axios.post(`/conversations/${conversationId}/messages`, {
                        type: "photo",
                        photo: base64Photo
                    });
                    
                    this.clearPhoto();
                    await this.refresh();
                    this.loading = false;
                };
                reader.onerror = () => {
                    this.errormsg = "Failed to read image file";
                    this.loading = false;
                };
                reader.readAsDataURL(this.selectedPhoto);
            } catch (e) {
                this.errormsg = e.toString();
                this.loading = false;
            }
        },
        
        showContextMenu(event, message) {
            event.preventDefault();
            this.contextMenu.show = true;
            this.contextMenu.x = event.clientX;
            this.contextMenu.y = event.clientY;
            this.contextMenu.messageId = message.id;
            this.contextMenu.isOwnMessage = message.sender === this.currentUsername;
        },
        
        hideContextMenu() {
            this.contextMenu.show = false;
        },
        
        async deleteMessage(messageId) {
            if (!confirm('Delete this message?')) return;
            
            this.hideContextMenu();
            this.errormsg = null;
            try {
                const conversationId = this.$route.params.conversationId;
                await this.$axios.delete(`/conversations/${conversationId}/messages/${messageId}`);
                await this.refresh();
            } catch (e) {
                this.errormsg = e.toString();
            }
        },
        
        async loadConversationsForForward() {
            try {
                let response = await this.$axios.get("/conversations");
                const currentConvoId = this.$route.params.conversationId;
                this.availableConversations = (response.data || []).filter(
                    conv => conv.id != currentConvoId
                );
            } catch (e) {
                console.error('Error loading conversations:', e);
            }
        },
        
        // Group settings handlers
        openGroupSettings() {
            this.showGroupSettings = true
            this.editingGroupName = false
            this.groupPhotoPreview = null
            this.groupSelectedPhoto = null
            this.newGroupMember = ''
            this.groupMemberSearchQuery = ''
            this.groupMemberSearchResults = []
            this.newGroupName = this.conversation?.name || ''
        },
        closeGroupSettings() {
            this.showGroupSettings = false
        },
        onGroupPhotoSelected(e) {
            const file = e.target.files[0]
            if (!file) return
            if (!file.type.startsWith('image/')) {
                this.errormsg = 'Please select an image file'
                return
            }
            this.groupSelectedPhoto = file
            const reader = new FileReader()
            reader.onload = (ev) => { this.groupPhotoPreview = ev.target.result }
            reader.readAsDataURL(file)
        },
        async saveGroupName() {
            if (!this.newGroupName.trim()) return
            const groupId = this.$route.params.conversationId
            this.loading = true
            this.errormsg = null
            try {
                await this.$axios.put(`/groups/${groupId}/name`, { name: this.newGroupName.trim() })
                await this.refresh()
                this.editingGroupName = false
            } catch (e) {
                this.errormsg = e.toString()
            }
            this.loading = false
        },
        async uploadGroupPhoto() {
            if (!this.groupPhotoPreview) return
            const groupId = this.$route.params.conversationId
            this.loading = true
            this.errormsg = null
            try {
                const base64 = this.groupPhotoPreview.split(',')[1]
                await this.$axios.put(`/groups/${groupId}/photo`, { photo: base64 })
                await this.refresh()
                this.groupPhotoPreview = null
                this.groupSelectedPhoto = null
                this.$refs.groupPhotoInput && (this.$refs.groupPhotoInput.value = '')
            } catch (e) {
                this.errormsg = e.toString()
            }
            this.loading = false
        },
        async searchGroupMembers() {
            const q = (this.groupMemberSearchQuery || '').trim()
            this.groupMemberSearchResults = []
            if (!q) return
            this.groupMemberSearchLoading = true
            try {
                const res = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(q)}`)
                this.groupMemberSearchResults = res.data || []
            } catch (e) {
                console.error('Member search failed:', e)
            }
            this.groupMemberSearchLoading = false
        },
        async addGroupMemberByUsername(username) {
            const u = username.trim()
            if (!u) return
            const groupId = this.$route.params.conversationId
            this.loading = true
            this.errormsg = null
            try {
                await this.$axios.post(`/groups/${groupId}/members`, { members: [u] })
                this.groupMemberSearchQuery = ''
                this.groupMemberSearchResults = []
                
                // update group data to get new members list
                const groupRes = await this.$axios.get(`/groups/${groupId}`)
                if (groupRes.data?.members) {
                    // Update conversation participants with the new members list
                    this.conversation.participants = groupRes.data.members
                }
                
                await this.refresh()
            } catch (e) {
                this.errormsg = e.toString()
            }
            this.loading = false
        },
        async leaveGroup() {
            if (!confirm('Leave this group?')) return
            const groupId = this.$route.params.conversationId
            this.loading = true
            this.errormsg = null
            try {
                await this.$axios.delete(`/groups/${groupId}/members`)
                this.$router.push('/home')
            } catch (e) {
                this.errormsg = e.toString()
            }
            this.loading = false
        },

        showForwardMenu(messageId) {
            this.forwardMessageId = messageId;
            this.showForwardDialog = true;
            this.hideContextMenu();
            this.loadConversationsForForward();
            this.forwardSearchQuery = '';
            this.forwardSearchResults = [];
        },
        async forwardMessage() {
            if (!this.selectedConversationId) {
                this.errormsg = "Please select a conversation";
                return;
            }

            this.loading = true;
            this.errormsg = null;
            try {
                const msg = this.messages.find(m => m.id === this.forwardMessageId)
                if (!msg) throw new Error('Message not found')

                const targetId = this.selectedConversationId
                if (msg.photo) {
                    await this.$axios.post(`/conversations/${targetId}/messages`, {
                        type: 'photo',
                        photo: msg.photo
                    })
                } else if (msg.text) {
                    await this.$axios.post(`/conversations/${targetId}/messages`, {
                        type: 'text',
                        text: msg.text
                    })
                } else {
                    throw new Error('Unsupported message type')
                }

                alert('Message forwarded successfully!');
                this.showForwardDialog = false;
                this.selectedConversationId = null;
            } catch (e) {
                this.errormsg = e.toString();
            }
            this.loading = false;
        },

        async searchForwardUsers() {
            const q = (this.forwardSearchQuery || '').trim();
            this.forwardSearchResults = [];
            if (!q) return;
            this.forwardSearchLoading = true;
            try {
                const res = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(q)}`);
                this.forwardSearchResults = res.data || [];
            } catch (e) {
                console.error('Forward user search failed:', e);
            }
            this.forwardSearchLoading = false;
        },
        async forwardToUser(username) {
            if (!username) return;
            this.loading = true;
            this.errormsg = null;
            try {
                const convoRes = await this.$axios.post('/conversations', { recipient: username });
                const targetId = convoRes.data.id;

                const msg = this.messages.find(m => m.id === this.forwardMessageId)
                if (!msg) throw new Error('Message not found')
                if (msg.photo) {
                    await this.$axios.post(`/conversations/${targetId}/messages`, { type: 'photo', photo: msg.photo })
                } else if (msg.text) {
                    await this.$axios.post(`/conversations/${targetId}/messages`, { type: 'text', text: msg.text })
                } else {
                    throw new Error('Unsupported message type')
                }

                alert('Message forwarded successfully!');
                this.showForwardDialog = false;
                this.selectedConversationId = null;
            } catch (e) {
                this.errormsg = e.toString();
            }
            this.loading = false;
        },

        
        toggleComments(messageId) {
            this.showComments[messageId] = !this.showComments[messageId];
            
            if (this.showComments[messageId] && !this.messageComments[messageId]) {
                this.fetchComments(messageId);
            }
        },
        
        async fetchComments(messageId) {
            try {
                const conversationId = this.$route.params.conversationId;
                let response = await this.$axios.get(`/conversations/${conversationId}/messages/${messageId}/comments`);
                this.messageComments[messageId] = response.data || [];
            } catch (e) {
                console.error('Error fetching comments:', e);
                this.messageComments[messageId] = [];
            }
        },
        
        async addComment(messageId) {
            if (!this.newComment[messageId]?.trim()) return;
            
            this.errormsg = null;
            try {
                const conversationId = this.$route.params.conversationId;
                await this.$axios.post(`/conversations/${conversationId}/messages/${messageId}/comments`, {
                    text: this.newComment[messageId]
                });
                this.newComment[messageId] = '';
                await this.fetchComments(messageId);
                await this.refresh();
            } catch (e) {
                this.errormsg = e.toString();
            }
        },
        
        async deleteComment(messageId, commentId) {
            if (!confirm('Delete this comment?')) return;
            
            this.errormsg = null;
            try {
                const conversationId = this.$route.params.conversationId;
                await this.$axios.delete(`/conversations/${conversationId}/messages/${messageId}/comments/${commentId}`);
                await this.fetchComments(messageId);
                await this.refresh();
            } catch (e) {
                this.errormsg = e.toString();
            }
        }
    },
    mounted() {
        this.refresh();
        document.addEventListener('click', this.hideContextMenu);
    },
    beforeUnmount() {
        document.removeEventListener('click', this.hideContextMenu);
    }
}
</script>

<template>
    <div class="chat-container">
        <!-- Chat Header -->
        <div class="chat-header">
            <div class="header-info">
                <div class="avatar-container">
                    <img 
                        v-if="conversation?.type === 'group' && conversation?.convo_pic"
                        :src="`data:image/png;base64,${conversation.convo_pic}`" 
                        class="chat-avatar"
                        alt="Group"
                    />
                    <img 
                        v-else-if="recipientPhoto" 
                        :src="`data:image/png;base64,${recipientPhoto}`" 
                        class="chat-avatar"
                        alt="Profile"
                    />
                    <div v-else class="avatar-placeholder">
                        {{ conversation?.type === 'group' ? '👥' : '👤' }}
                    </div>
                </div>
                
                <h1 class="chat-title">
                    {{ conversation?.type === 'group' ? (conversation?.name || 'Group') : (getOtherParticipant() || 'Chat') }}
                </h1>
            </div>
            
            <div class="header-actions">
                <button v-if="conversation?.type === 'group'" type="button" class="btn-header" @click="openGroupSettings" title="Group Settings">
                    ⚙️
                </button>
                <button type="button" class="btn-header" @click="refresh" title="Refresh">
                    🔄
                </button>
                <button type="button" class="btn-header" @click="$router.push('/home')" title="Back">
                    ⬅️
                </button>
            </div>
        </div>

        <ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

        <!-- Messages Area -->
        <div class="messages-area">
            <div 
                v-for="msg in messages" 
                :key="msg.id" 
                class="message-wrapper"
                :class="{ 'own-message': msg.sender === currentUsername }"
                @contextmenu="showContextMenu($event, msg)"
            >
                <div class="message-bubble">
                    <div class="message-sender">{{ msg.sender }}</div>
                    <p v-if="msg.text" class="message-text">{{ msg.text }}</p>
                    <img v-if="msg.photo" :src="`data:image/png;base64,${msg.photo}`" class="message-image" />
                    
                    <div class="message-meta">
                        <span class="message-time">{{ new Date(msg.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) }}</span>
                        <span v-if="msg.sender === currentUsername" class="message-status">✓</span>
                        <span v-else class="message-status">✓✓</span>
                    </div>
                    
                    <div class="message-actions">
                        <button @click="toggleComments(msg.id)" class="btn-comment">
                            💬 {{ msg.comments_count || msg.comments_authors?.length || 0 }}
                        </button>
                    </div>
                    
                    <!-- Comments Section -->
                    <div v-if="showComments[msg.id]" class="comments-section">
                        <div class="comments-header">Comments</div>
                        
                        <div v-if="messageComments[msg.id] && messageComments[msg.id].length > 0" class="comments-list">
                            <div v-for="comment in messageComments[msg.id]" :key="comment.id" class="comment-item">
                                <div class="comment-header">
                                    <strong>{{ comment.username }}</strong>
                                    <button 
                                        v-if="comment.username === currentUsername" 
                                        @click="deleteComment(msg.id, comment.id)"  
                                        class="btn-delete-comment"
                                    >
                                        🗑️
                                    </button>
                                </div>
                                <p class="comment-text">{{ comment.text }}</p>
                            </div>
                        </div>
                        
                        <div v-else class="no-comments">No comments yet</div>
                        
                        <div class="comment-input-wrapper">
                            <input 
                                v-model="newComment[msg.id]" 
                                type="text" 
                                class="comment-input" 
                                placeholder="Write a comment..."
                                @keyup.enter="addComment(msg.id)"
                            />
                            <button @click="addComment(msg.id)" class="btn-add-comment">➤</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Context Menu -->
        <div 
            v-if="contextMenu.show" 
            class="context-menu"
            :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }"
            @click.stop
        >
            <button 
                v-if="contextMenu.isOwnMessage"
                @click="deleteMessage(contextMenu.messageId)" 
                class="context-item danger"
            >
                🗑️ Delete
            </button>
            <button 
                @click="showForwardMenu(contextMenu.messageId)" 
                class="context-item"
            >
                ➡️ Forward
            </button>
        </div>

        <!-- Forward Dialog -->
        <div v-if="showForwardDialog" class="modal-overlay" @click="showForwardDialog = false">
            <div class="modal-dialog" @click.stop>
                <div class="modal-header">
                    <h3>✉️ Forward Message</h3>
                    <button @click="showForwardDialog = false" class="btn-modal-close">✕</button>
                </div>
                <div class="modal-body">
                    <label class="modal-label">Search Users</label>
                    <div class="search-group">
                        <input 
                            v-model="forwardSearchQuery" 
                            @keyup.enter="searchForwardUsers" 
                            type="text" 
                            class="modal-input" 
                            placeholder="Type username..." 
                        />
                        <button class="btn-modal-search" @click="searchForwardUsers" :disabled="forwardSearchLoading">
                            🔍
                        </button>
                    </div>
                    
                    <div v-if="forwardSearchResults && forwardSearchResults.length" class="search-results">
                        <div v-for="u in forwardSearchResults" :key="u.username" class="result-item">
                            <div class="result-user">
                                <img v-if="u.pic" :src="`data:image/png;base64,${u.pic}`" class="result-avatar" />
                                <div v-else class="result-avatar-placeholder">{{ u.username[0].toUpperCase() }}</div>
                                <span>{{ u.username }}</span>
                            </div>
                            <button class="btn-result-action" @click="forwardToUser(u.username)">Send</button>
                        </div>
                    </div>

                    <div class="divider">or select a conversation</div>
                    
                    <div class="conversation-list">
                        <label 
                            v-for="conv in availableConversations" 
                            :key="conv.id"
                            class="conversation-item"
                        >
                            <input 
                                type="radio" 
                                :value="conv.id" 
                                v-model="selectedConversationId"
                                class="conversation-radio"
                            />
                            <span>{{ conv.name || (conv.participants?.filter(p => p !== currentUsername)?.join(', ')) || 'Unknown' }}</span>
                        </label>
                        <div v-if="!availableConversations.length" class="empty-list">
                            No conversations available
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button @click="showForwardDialog = false" class="btn-modal-cancel">Cancel</button>
                    <button 
                        @click="forwardMessage" 
                        class="btn-modal-confirm"
                        :disabled="!selectedConversationId || loading"
                    >
                        Forward
                    </button>
                </div>
            </div>
        </div>

        <!-- Group Settings Dialog -->
        <div v-if="showGroupSettings" class="modal-overlay" @click="closeGroupSettings">
            <div class="modal-dialog modal-large" @click.stop>
                <div class="modal-header">
                    <h3>⚙️ Group Settings</h3>
                    <button class="btn-modal-close" @click="closeGroupSettings">✕</button>
                </div>
                <div class="modal-body">
                    <!-- Group Name -->
                    <div class="settings-section">
                        <label class="modal-label">Group Name</label>
                        <div class="input-with-button">
                            <input v-model="newGroupName" type="text" class="modal-input" placeholder="Enter group name" />
                            <button class="btn-modal-action" @click="saveGroupName">💾 Save</button>
                        </div>
                    </div>
                    
                    <!-- Group Photo -->
                    <div class="settings-section">
                        <label class="modal-label">Group Photo</label>
                        <div class="photo-upload-section">
                            <input ref="groupPhotoInput" type="file" class="file-input-hidden" accept="image/*" @change="onGroupPhotoSelected" id="groupPhotoUpload" />
                            <label for="groupPhotoUpload" class="btn-file-select">📷 Choose Photo</label>
                            <button class="btn-modal-action" @click="uploadGroupPhoto" :disabled="!groupPhotoPreview">⬆️ Upload</button>
                        </div>
                        <div v-if="groupPhotoPreview" class="photo-preview-small">
                            <img :src="groupPhotoPreview" alt="Preview" />
                        </div>
                    </div>
                    
                    <!-- Current Members -->
                    <div class="settings-section">
                        <label class="modal-label">Current Members ({{ conversation?.participants?.length || 0 }})</label>
                        <div class="members-list">
                            <div v-for="member in conversation?.participants" :key="member" class="member-item">
                                <span class="member-emoji">👤</span>
                                <span class="member-name">{{ member }}</span>
                            </div>
                        </div>
                    </div>
                    
                    <!-- Add Member -->
                    <div class="settings-section">
                        <label class="modal-label">Add New Member</label>
                        <div class="search-group">
                            <input 
                                v-model="groupMemberSearchQuery" 
                                @keyup.enter="searchGroupMembers" 
                                type="text" 
                                class="modal-input" 
                                placeholder="Search username..." 
                            />
                            <button class="btn-modal-search" @click="searchGroupMembers" :disabled="groupMemberSearchLoading">
                                🔍
                            </button>
                        </div>
                        
                        <div v-if="groupMemberSearchResults && groupMemberSearchResults.length" class="search-results">
                            <div v-for="u in groupMemberSearchResults" :key="u.username" class="result-item">
                                <div class="result-user">
                                    <img v-if="u.pic" :src="`data:image/png;base64,${u.pic}`" class="result-avatar" />
                                    <div v-else class="result-avatar-placeholder">{{ u.username[0].toUpperCase() }}</div>
                                    <span>{{ u.username }}</span>
                                </div>
                                <button class="btn-result-action" @click="addGroupMemberByUsername(u.username)">➕ Add</button>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button class="btn-modal-danger" @click="leaveGroup">🚪 Leave Group</button>
                    <button class="btn-modal-cancel" @click="closeGroupSettings">Close</button>
                </div>
            </div>
        </div>


        <!-- Photo Preview -->
        <div v-if="photoPreview" class="photo-preview-card">
            <div class="preview-header">
                <span>📸 Photo Selected</span>
                <button @click="clearPhoto" class="btn-clear-photo">✕</button>
            </div>
            <img :src="photoPreview" class="preview-image" />
        </div>

        <!-- Message Input -->
        <div class="message-input-container">
            <input 
                v-model="newMessage" 
                type="text" 
                class="message-input" 
                placeholder="Type a message..."
                @keyup.enter="sendMessage"
                :disabled="loading || photoPreview"
            />
            <input 
                type="file" 
                ref="photoInput"
                @change="handleFileSelect"
                accept="image/*"
                class="file-input-hidden"
                id="messagePhoto"
            />
            <label for="messagePhoto" class="btn-attach" :class="{ 'disabled': loading }">
                📷
            </label>
            <button @click="sendMessage" class="btn-send" :disabled="loading">
                {{ photoPreview ? '📤' : '➤' }}
            </button>
        </div>
    </div>
</template>

<style scoped>
.chat-container {
    max-width: 900px;
    margin: 1.5rem auto;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    height: calc(100vh - 3rem);
}

.chat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    background: white;
    border-radius: 20px 20px 0 0;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
    margin-bottom: 0;
}

.header-info {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.avatar-container {
    position: relative;
}

.chat-avatar {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    object-fit: cover;
    border: 3px solid #e8f0fe;
}

.avatar-placeholder {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.5rem;
    border: 3px solid #e8f0fe;
}

.chat-title {
    font-size: 1.4rem;
    font-weight: 600;
    color: #2d3748;
    margin: 0;
}

.header-actions {
    display: flex;
    gap: 0.5rem;
}

.btn-header {
    background: white;
    border: 2px solid #e2e8f0;
    padding: 0.5rem 0.75rem;
    border-radius: 12px;
    font-size: 1.1rem;
    cursor: pointer;
    transition: all 0.3s ease;
}

.btn-header:hover {
    background: #f7fafc;
    transform: translateY(-2px);
}



.messages-area {
    flex: 1;
    overflow-y: auto;
    padding: 1.5rem;
    background: #f7fafc;
    display: flex;
    flex-direction: column-reverse;
    gap: 1rem;
}

.message-wrapper {
    display: flex;
    justify-content: flex-start;
}

.message-wrapper.own-message {
    justify-content: flex-end;
}

.message-bubble {
    max-width: 70%;
    background: white;
    padding: 1rem 1.25rem;
    border-radius: 18px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    transition: all 0.3s ease;
}

.own-message .message-bubble {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
}

.message-bubble:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.message-sender {
    font-size: 0.85rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
    opacity: 0.8;
}

.own-message .message-sender {
    color: rgba(255, 255, 255, 0.9);
}

.message-text {
    margin: 0 0 0.5rem 0;
    line-height: 1.5;
}

.message-image {
    max-width: 100%;
    border-radius: 12px;
    margin: 0.5rem 0;
}

.message-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.75rem;
    opacity: 0.7;
    margin-top: 0.5rem;
}

.message-time {
    font-size: 0.75rem;
}

.message-status {
    font-size: 0.9rem;
}

.message-actions {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid rgba(0, 0, 0, 0.1);
}

.own-message .message-actions {
    border-top: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-comment {
    background: transparent;
    border: 1px solid rgba(0, 0, 0, 0.2);
    padding: 0.4rem 0.75rem;
    border-radius: 12px;
    font-size: 0.85rem;
    cursor: pointer;
    transition: all 0.3s ease;
}

.own-message .btn-comment {
    border-color: rgba(255, 255, 255, 0.4);
    color: white;
}

.btn-comment:hover {
    background: rgba(0, 0, 0, 0.05);
}

.comments-section {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 2px solid rgba(0, 0, 0, 0.1);
}

.own-message .comments-section {
    border-top-color: rgba(255, 255, 255, 0.2);
}

.comments-header {
    font-weight: 600;
    margin-bottom: 0.75rem;
    font-size: 0.9rem;
}

.comments-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1rem;
}

.comment-item {
    background: rgba(0, 0, 0, 0.05);
    padding: 0.75rem;
    border-radius: 10px;
}

.own-message .comment-item {
    background: rgba(255, 255, 255, 0.15);
}

.comment-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.25rem;
    font-size: 0.85rem;
}

.comment-text {
    margin: 0;
    font-size: 0.9rem;
}

.btn-delete-comment {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 0.9rem;
    opacity: 0.7;
    transition: opacity 0.3s ease;
}

.btn-delete-comment:hover {
    opacity: 1;
}

.no-comments {
    font-size: 0.85rem;
    opacity: 0.7;
    margin-bottom: 1rem;
    font-style: italic;
}

.comment-input-wrapper {
    display: flex;
    gap: 0.5rem;
}

.comment-input {
    flex: 1;
    padding: 0.6rem 1rem;
    border: 1px solid rgba(0, 0, 0, 0.2);
    border-radius: 12px;
    font-size: 0.9rem;
    background: rgba(255, 255, 255, 0.9);
}

.own-message .comment-input {
    background: rgba(255, 255, 255, 0.95);
    color: #2d3748;
}

.comment-input:focus {
    outline: none;
    border-color: #667eea;
}

.btn-add-comment {
    background: #667eea;
    color: white;
    border: none;
    padding: 0.6rem 1rem;
    border-radius: 12px;
    cursor: pointer;
    font-size: 1rem;
    transition: all 0.3s ease;
}

.btn-add-comment:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(102, 126, 234, 0.3);
}

.context-menu {
    position: fixed;
    background: white;
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    z-index: 3000;
    overflow: hidden;
}

.context-item {
    display: block;
    width: 100%;
    padding: 0.75rem 1.25rem;
    border: none;
    background: none;
    text-align: left;
    cursor: pointer;
    color: #2d3748;
    font-size: 0.95rem;
    transition: all 0.2s ease;
}

.context-item:hover {
    background: #f7fafc;
}

.context-item.danger {
    color: #e53e3e;
}

.context-item.danger:hover {
    background: #fff5f5;
}

.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 5000;
    padding: 1rem;
    pointer-events: auto;
}

.modal-dialog {
    background: white;
    border-radius: 20px;
    width: 100%;
    max-width: 500px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
    pointer-events: auto;
    position: relative;
    z-index: 5001;
}

.modal-large {
    max-width: 600px;
}

.modal-header {
    padding: 1.5rem;
    border-bottom: 2px solid #f0f0f0;
    display: flex;
    justify-content: space-between;
    align-items: center;
    pointer-events: auto;
}

.modal-header h3 {
    margin: 0;
    font-size: 1.3rem;
    color: #2d3748;
}

.btn-modal-close {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: #718096;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-modal-close:hover {
    background: #f7fafc;
    color: #2d3748;
}

.modal-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
    pointer-events: auto;
}

.modal-label {
    display: block;
    font-weight: 600;
    color: #4a5568;
    margin-bottom: 0.75rem;
    font-size: 0.95rem;
}

.search-group {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.modal-input {
    flex: 1;
    padding: 0.75rem 1rem;
    border: 2px solid #e2e8f0;
    border-radius: 12px;
    font-size: 0.95rem;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.modal-input:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.btn-modal-search {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    padding: 0.75rem 1.25rem;
    border-radius: 12px;
    cursor: pointer;
    font-size: 1.1rem;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-modal-search:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-modal-search:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.search-results {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.result-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: #f7fafc;
    border-radius: 12px;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.result-item:hover {
    background: #edf2f7;
}

.result-user {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.result-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid #e8f0fe;
}

.result-avatar-placeholder {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    border: 2px solid #e8f0fe;
}

.btn-result-action {
    background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 12px;
    cursor: pointer;
    font-size: 0.85rem;
    font-weight: 500;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-result-action:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 10px rgba(79, 172, 254, 0.3);
}

.divider {
    text-align: center;
    color: #718096;
    font-size: 0.85rem;
    margin: 1.5rem 0;
    position: relative;
}

.divider::before,
.divider::after {
    content: '';
    position: absolute;
    top: 50%;
    width: 40%;
    height: 1px;
    background: #e2e8f0;
}

.divider::before {
    left: 0;
}

.divider::after {
    right: 0;
}

.conversation-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 200px;
    overflow-y: auto;
}

.conversation-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    background: #f7fafc;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.conversation-item:hover {
    background: #edf2f7;
}

.conversation-radio {
    cursor: pointer;
    pointer-events: auto;
}

.empty-list {
    text-align: center;
    color: #718096;
    padding: 2rem;
    font-style: italic;
}

.modal-footer {
    padding: 1.5rem;
    border-top: 2px solid #f0f0f0;
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    pointer-events: auto;
}

.btn-modal-cancel {
    background: #f7fafc;
    color: #718096;
    border: 2px solid #e2e8f0;
    padding: 0.6rem 1.5rem;
    border-radius: 12px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-modal-cancel:hover {
    background: #edf2f7;
    border-color: #cbd5e0;
}

.btn-modal-confirm {
    background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
    color: white;
    border: none;
    padding: 0.6rem 1.5rem;
    border-radius: 12px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-modal-confirm:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(56, 239, 125, 0.3);
}

.btn-modal-confirm:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.btn-modal-danger {
    background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
    color: white;
    border: none;
    padding: 0.6rem 1.5rem;
    border-radius: 12px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-modal-danger:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(238, 90, 111, 0.3);
}

.settings-section {
    margin-bottom: 2rem;
}

.input-with-button {
    display: flex;
    gap: 0.5rem;
}

.btn-modal-action {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    padding: 0.6rem 1.25rem;
    border-radius: 12px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    white-space: nowrap;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-modal-action:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-modal-action:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.photo-upload-section {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}

.file-input-hidden {
    display: none;
}

.btn-file-select {
    background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
    color: white;
    padding: 0.6rem 1.25rem;
    border-radius: 12px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.btn-file-select:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(79, 172, 254, 0.3);
}

.photo-preview-small {
    margin-top: 1rem;
}

.photo-preview-small img {
    max-height: 120px;
    border-radius: 12px;
    border: 3px solid #e8f0fe;
}

.members-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 150px;
    overflow-y: auto;
    padding: 0.5rem;
    background: #f7fafc;
    border-radius: 12px;
}

.member-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0.75rem;
    background: white;
    border-radius: 8px;
}

.member-emoji {
    font-size: 1.2rem;
}

.member-name {
    font-weight: 500;
    color: #2d3748;
}

.photo-preview-card {
    background: white;
    border-radius: 16px;
    padding: 1rem;
    margin-bottom: 1rem;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.preview-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75rem;
    font-weight: 500;
}

.btn-clear-photo {
    background: #ff6b6b;
    color: white;
    border: none;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s ease;
}

.btn-clear-photo:hover {
    transform: scale(1.1);
}

.preview-image {
    max-width: 100%;
    max-height: 200px;
    border-radius: 12px;
    object-fit: contain;
}

.message-input-container {
    display: flex;
    gap: 0.75rem;
    padding: 1rem 1.5rem;
    background: white;
    border-radius: 0 0 20px 20px;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.message-input {
    flex: 1;
    padding: 0.75rem 1.25rem;
    border: 2px solid #e2e8f0;
    border-radius: 20px;
    font-size: 1rem;
    transition: all 0.3s ease;
}

.message-input:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.message-input:disabled {
    background: #f7fafc;
    cursor: not-allowed;
}

.btn-attach {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    padding: 0.75rem 1.25rem;
    border-radius: 20px;
    cursor: pointer;
    font-size: 1.2rem;
    transition: all 0.3s ease;
    display: flex;
    align-items: center;
    justify-content: center;
}

.btn-attach:hover:not(.disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-attach.disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.btn-send {
    background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 20px;
    cursor: pointer;
    font-size: 1.2rem;
    font-weight: 600;
    transition: all 0.3s ease;
}

.btn-send:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(56, 239, 125, 0.3);
}

.btn-send:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

@media (max-width: 768px) {
    .chat-container {
        padding: 0.5rem;
        height: 100vh;
        margin: 0;
    }
    
    .chat-header {
        border-radius: 0;
        padding: 1rem;
    }
    
    .chat-title {
        font-size: 1.1rem;
    }
    
    .message-bubble {
        max-width: 85%;
    }
    
    .message-input-container {
        border-radius: 0;
    }
    
    .modal-dialog {
        max-width: 95%;
        max-height: 90vh;
    }
}
</style>
