<script>
export default {
    props: {
        show: {
            type: Boolean,
            required: true
        }
    },
    emits: ['close', 'chat-created'],
    data() {
        return {
            chatSearchQuery: '',
            chatSearchLoading: false,
            chatSearchResults: [],
            selectedChatUser: null,
            creatingChat: false,
            errormsg: null
        }
    },
    methods: {
        close() {
            this.$emit('close');
            this.resetForm();
        },
        resetForm() {
            this.chatSearchQuery = '';
            this.chatSearchResults = [];
            this.selectedChatUser = null;
            this.errormsg = null;
        },
        async searchChatUsers() {
            const q = (this.chatSearchQuery || '').trim();
            this.chatSearchResults = [];
            if (!q) return;
            this.chatSearchLoading = true;
            try {
                const res = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(q)}`);
                this.chatSearchResults = res.data || [];
            } catch (e) {
                console.error('Chat user search failed:', e);
            }
            this.chatSearchLoading = false;
        },
        pickChatUser(username) {
            this.selectedChatUser = username;
        },
        async createChatConfirm() {
            if (!this.selectedChatUser) return;
            this.creatingChat = true;
            this.errormsg = null;
            try {
                let response = await this.$axios.post('/conversations', { 
                    recipient: this.selectedChatUser 
                });
                this.$emit('chat-created', response.data.id);
                this.close();
            } catch (e) {
                this.errormsg = e.toString();
            }
            this.creatingChat = false;
        }
    }
}
</script>

<template>
    <div v-if="show" class="modal-overlay" @click="close">
        <div class="modal-dialog" @click.stop>
            <div class="modal-header">
                <h3>✉️ Start New Chat</h3>
                <button class="btn-modal-close" @click="close">✕</button>
            </div>
            <div class="modal-body">
                <ErrorMsg v-if="errormsg" :msg="errormsg" />
                
                <div class="settings-section">
                    <label class="modal-label">Search User</label>
                    <div class="search-group">
                        <input 
                            v-model="chatSearchQuery" 
                            @keyup.enter="searchChatUsers" 
                            type="text" 
                            class="modal-input" 
                            placeholder="Type username..." 
                        />
                        <button class="btn-modal-search" @click="searchChatUsers" :disabled="chatSearchLoading">
                            🔍
                        </button>
                    </div>
                </div>
                
                <div v-if="chatSearchResults && chatSearchResults.length" class="search-results-chat">
                    <div class="results-label">Select a user to chat with:</div>
                    <button 
                        v-for="u in chatSearchResults" 
                        :key="u.username"
                        class="result-item-btn"
                        :class="{ 'selected': selectedChatUser === u.username }"
                        @click="pickChatUser(u.username)"
                    >
                        <div class="result-user">
                            <img v-if="u.pic" :src="`data:image/png;base64,${u.pic}`" class="result-avatar" />
                            <div v-else class="result-avatar-placeholder">{{ u.username[0].toUpperCase() }}</div>
                            <span>{{ u.username }}</span>
                        </div>
                        <span v-if="selectedChatUser === u.username" class="selected-badge">✓ Selected</span>
                    </button>
                </div>
                <div v-else-if="!chatSearchLoading" class="search-hint">
                    <span class="hint-emoji">🔍</span>
                    <p>Type a username to search</p>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-modal-cancel" @click="close">Cancel</button>
                <button 
                    class="btn-cute btn-primary" 
                    :disabled="creatingChat || !selectedChatUser" 
                    @click="createChatConfirm"
                >
                    <span v-if="creatingChat" class="spinner-sm"></span>
                    {{ creatingChat ? 'Starting...' : '💬 Start Chat' }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.search-results-chat {
    margin-top: 1rem;
}

.results-label {
    font-size: 0.9rem;
    color: var(--text-secondary);
    margin-bottom: 0.75rem;
    font-weight: 500;
}

.search-hint {
    text-align: center;
    padding: 3rem 2rem;
    color: var(--text-secondary);
}

.hint-emoji {
    font-size: 3rem;
    display: block;
    margin-bottom: 1rem;
    opacity: 0.5;
}

.search-hint p {
    margin: 0;
    font-style: italic;
}

.result-item-btn {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: var(--bg-light);
    border-radius: var(--radius-sm);
    border: 2px solid transparent;
    cursor: pointer;
    transition: all 0.3s ease;
    pointer-events: auto;
}

.result-item-btn:hover {
    background: #edf2f7;
}

.result-item-btn.selected {
    background: #e8f0fe;
    border-color: #667eea;
}

.selected-badge {
    background: var(--success-gradient);
    color: white;
    padding: 0.35rem 0.85rem;
    border-radius: 15px;
    font-size: 0.85rem;
    font-weight: 600;
}
</style>
