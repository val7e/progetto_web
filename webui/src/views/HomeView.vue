<script>
import NewChatDialog from '@/components/NewChatDialog.vue'
import NewGroupDialog from '@/components/NewGroupDialog.vue'

export default {
    components: {
        NewChatDialog,
        NewGroupDialog
    },
    data: function() {
        return {
            errormsg: null,
            loading: false,
            conversations: [],
            currentUsername: localStorage.getItem('username'),
            userPhotos: {},
            showNewGroupDialog: false,
            showNewChatDialog: false
        }
    },
    methods: {
        async refresh() {
            this.loading = true;
            this.errormsg = null;
            try {
                let response = await this.$axios.get("/conversations");
                this.conversations = (response.data || []).slice();
                this.conversations.sort((a,b) => {
                    const ta = a.last_message?.timestamp ? new Date(a.last_message.timestamp).getTime() : 0;
                    const tb = b.last_message?.timestamp ? new Date(b.last_message.timestamp).getTime() : 0;
                    return tb - ta;
                });

                await this.fetchGroupSummaries();
                await this.fetchProfilePictures();
            } catch (e) {
                this.errormsg = e.toString();
            }
            this.loading = false;
        },
        
        async fetchGroupSummaries() {
    const filtered = []
    for (const conv of this.conversations) {
        if (conv.type !== 'group') { 
            filtered.push(conv); 
            continue 
        }
        try {
            const res = await this.$axios.get(`/groups/${conv.id}`)
            const members = res.data?.members || []
            
            // Check if current user is a member
            if (!members.includes(this.currentUsername)) {
                continue // Skip this group
            }
            
            // Update group info
            conv.name = res.data?.name || conv.name || 'Group'
            if (res.data?.group_photo) {
                conv.convo_pic = res.data.group_photo
            }
            
            filtered.push(conv)
        } catch (e) {
            filtered.push(conv) // Keep it even if fetch fails
        }
    }
    this.conversations = filtered
}
,
        
        async fetchProfilePictures() {
            const allUsernames = new Set();
            this.conversations.forEach(conv => {
                if (conv.type === 'group') return;
                conv.participants?.forEach(p => {
                    if (p !== this.currentUsername) {
                        allUsernames.add(p);
                    }
                });
            });
            
            for (const username of allUsernames) {
                if (!this.userPhotos[username]) {
                    try {
                        let response = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(username)}`);
                        if (response.data && response.data.length > 0) {
                            const user = response.data.find(u => u.username === username);
                            if (user && user.pic) {
                                this.userPhotos[username] = user.pic;
                            }
                        }
                    } catch (e) {
                    }
                }
            }
        },
        
        getOtherParticipant(conversation) {
            const others = conversation.participants?.filter(p => p !== this.currentUsername);
            return others?.join(', ') || 'Unknown';
        },
        
        getOtherParticipantUsername(conversation) {
            const others = conversation.participants?.filter(p => p !== this.currentUsername);
            return others?.[0] || null;
        },

        getConversationTitle(conversation) {
            if (conversation.type === 'group') {
                return conversation.name || 'Group';
            }
            return conversation.name ?? this.getOtherParticipant(conversation);
        },

        getConversationPic(conversation) {
            if (conversation.convo_pic) return conversation.convo_pic;
            if (conversation.type === 'group') return null;
            const other = this.getOtherParticipantUsername(conversation);
            return other ? this.userPhotos[other] : null;
        },

        getInitials(name) {
            if (!name) return '?';
            return name.split(' ').map(n => n[0]).join('').toUpperCase().substring(0, 2);
        },
        
        isPhotoPreview(preview) {
            if (!preview) return false;
            const p = String(preview).trim();
            return /^\[?photo\]?/i.test(p);
        },
        
        openNewChatDialog() {
            this.showNewChatDialog = true;
        },
        
        openNewGroupDialog() {
            this.showNewGroupDialog = true;
        },
        
        handleChatCreated(conversationId) {
            this.$router.push(`/chat/${conversationId}`);
        },
        
        async handleGroupCreated(groupId) {
            console.log('Group created with ID:', groupId);
            
            // refresh the conversations list to show the new group
            await this.refresh();
            
            // delay
            await new Promise(resolve => setTimeout(resolve, 300));
            
            // Try to fetch the group details explicitly
            try {
                const g = await this.$axios.get(`/groups/${groupId}`);
                console.log('Group details fetched:', g.data);
                
                // Find the conversation in the list and update it
                const convIndex = this.conversations.findIndex(c => c.id === groupId);
                if (convIndex !== -1) {
                    if (g.data?.name) this.conversations[convIndex].name = g.data.name;
                    if (g.data?.group_photo) this.conversations[convIndex].convo_pic = g.data.group_photo;
                }
            } catch (e) {
                console.error('Failed to fetch group details:', e);
            }
            
            // show to the new group chat
            this.$router.push(`/chat/${groupId}`);
        }

    },
    mounted() {
        this.refresh();
    }
}
</script>

<template>
    <div class="conversations-container">
        <!-- Header -->
        <div class="conversations-header">
            <h1 class="conversations-title">💬 Conversations</h1>
            <div class="header-actions">
                <button type="button" class="btn-cute btn-primary" @click="refresh">
                    🔄 Refresh
                </button>
                <button type="button" class="btn-cute btn-secondary" @click="openNewChatDialog">
                    ✉️ New Chat
                </button>
                <button type="button" class="btn-cute btn-success" @click="openNewGroupDialog">
                    👥 New Group
                </button>
            </div>
        </div>

        <ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

        <!-- Loading State -->
        <div v-if="loading" class="loading-container">
            <div class="spinner-border text-primary" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
        </div>

        <!-- Conversations List -->
        <div v-else class="conversations-list">
            <div 
                v-for="conv in conversations" 
                :key="conv.id" 
                class="conversation-item"
                @click="$router.push(`/chat/${conv.id}`)"
            >
                <!-- Avatar -->
                <div class="conversation-avatar">
                    <img 
                        v-if="getConversationPic(conv)" 
                        :src="`data:image/png;base64,${getConversationPic(conv)}`" 
                        class="avatar-cute"
                        alt="Avatar"
                    />
                    <div v-else class="avatar-placeholder">
                        {{ getInitials(getConversationTitle(conv)) }}
                    </div>
                </div>
                
                <!-- Conversation Info -->
                <div class="conversation-content">
                    <div class="conversation-header-row">
                        <div class="conversation-name-row">
                            <h3 class="conversation-name">{{ getConversationTitle(conv) }}</h3>
                            <span v-if="conv.type === 'group'" class="group-badge">👥</span>
                        </div>
                        <span class="conversation-time">
                            {{ conv.last_message?.timestamp ? new Date(conv.last_message.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : '' }}
                        </span>
                    </div>
                    <p class="conversation-preview">
                        <template v-if="conv.last_message">
                            <template v-if="isPhotoPreview(conv.last_message.preview)">
                                📷 Photo
                            </template>
                            <template v-else>
                                {{ conv.last_message.preview }}
                            </template>
                        </template>
                        <template v-else>
                            <span class="no-messages">No messages yet</span>
                        </template>
                    </p>
                </div>
            </div>
        </div>

        <!-- Empty State -->
        <div v-if="!loading && conversations.length === 0" class="empty-state">
            <span class="empty-emoji">💬</span>
            <p class="empty-text">No conversations yet</p>
            <p class="empty-hint">Start a new chat to begin messaging!</p>
        </div>

        <!-- Dialog Components -->
        <NewChatDialog 
            :show="showNewChatDialog" 
            @close="showNewChatDialog = false"
            @chat-created="handleChatCreated"
        />
        
        <NewGroupDialog 
            :show="showNewGroupDialog" 
            @close="showNewGroupDialog = false"
            @group-created="handleGroupCreated"
        />
    </div>
</template>

<style scoped>
.conversations-container {
    max-width: 800px;
    margin: 2rem auto;
    padding: 1.5rem;
}

.conversations-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid #f0f0f0;
    flex-wrap: wrap;
    gap: 1rem;
}

.conversations-title {
    font-size: 1.75rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
}

.header-actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
}

.loading-container {
    text-align: center;
    padding: 3rem;
}

.conversations-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.conversation-item {
    background: white;
    border-radius: var(--radius-md);
    padding: 1rem 1.25rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    box-shadow: var(--shadow-sm);
    transition: all 0.3s ease;
    cursor: pointer;
}

.conversation-item:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
}

.conversation-avatar {
    flex-shrink: 0;
}

.conversation-content {
    flex: 1;
    min-width: 0;
}

.conversation-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
}

.conversation-name-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex: 1;
    min-width: 0;
}

.conversation-name {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.group-badge {
    font-size: 0.9rem;
    flex-shrink: 0;
}

.conversation-time {
    font-size: 0.8rem;
    color: var(--text-secondary);
    white-space: nowrap;
    flex-shrink: 0;
    margin-left: 0.5rem;
}

.conversation-preview {
    font-size: 0.95rem;
    color: var(--text-secondary);
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.no-messages {
    font-style: italic;
    opacity: 0.7;
}

.empty-state {
    text-align: center;
    padding: 4rem 2rem;
    background: white;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
}

.empty-emoji {
    font-size: 4rem;
    display: block;
    margin-bottom: 1rem;
}

.empty-text {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 0.5rem 0;
}

.empty-hint {
    font-size: 0.95rem;
    color: var(--text-secondary);
    margin: 0;
}

@media (max-width: 768px) {
    .conversations-container {
        padding: 1rem;
    }
    
    .conversations-header {
        flex-direction: column;
        align-items: stretch;
    }
    
    .header-actions {
        width: 100%;
    }
    
    .header-actions button {
        flex: 1;
    }
    
    .conversation-item {
        padding: 0.75rem 1rem;
    }
}
</style>
