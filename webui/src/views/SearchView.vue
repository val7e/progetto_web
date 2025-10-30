<script>
export default {
    data: function() {
        return {
            query: '',
            results: [],
            loading: false,
            errormsg: null
        }
    },
    methods: {
        async search() {
            if (!this.query.trim()) {
                this.results = []
                return
            }
            this.loading = true
            this.errormsg = null
            try {
                const res = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(this.query.trim())}`)
                this.results = res.data || []
            } catch (e) {
                this.errormsg = e.toString()
            }
            this.loading = false
        },
        async startChat(username) {
            try {
                const res = await this.$axios.post('/conversations', {
                    recipient: username
                })
                this.$router.push(`/chat/${res.data.id}`)
            } catch (e) {
                this.errormsg = e.toString()
            }
        },
        initials(name) {
            if (!name) return '?'
            return name.split(' ').map(n => n[0]).join('').toUpperCase().substring(0, 2)
        }
    }
}
</script>

<template>
    <div class="search-container">
        <div class="search-header">
            <h1 class="search-title">🔍 Search Users</h1>
        </div>

        <ErrorMsg v-if="errormsg" :msg="errormsg" />

        <div class="search-box">
            <input 
                v-model="query" 
                @keyup.enter="search" 
                type="text" 
                class="search-input" 
                placeholder="Type a username..." 
            />
            <button class="btn-cute btn-primary" @click="search" :disabled="loading">
                <span v-if="loading" class="spinner"></span>
                <span v-else>🔎</span>
            </button>
        </div>

        <div v-if="results.length > 0" class="results-container">
            <p class="results-count">Found {{ results.length }} user{{ results.length !== 1 ? 's' : '' }}</p>
            
            <div class="user-cards">
                <div 
                    v-for="user in results" 
                    :key="user.username" 
                    class="user-card"
                >
                    <div class="user-info">
                        <div class="avatar-wrapper">
                            <img 
                                v-if="user.pic" 
                                :src="`data:image/png;base64,${user.pic}`" 
                                class="user-avatar"
                                alt="Avatar"
                            />
                            <div v-else class="avatar-initials">
                                {{ initials(user.username) }}
                            </div>
                        </div>
                        <span class="username">{{ user.username }}</span>
                    </div>
                    <!-- Chat button -->
                    <button class="btn-cute btn-secondary" @click="startChat(user.username)">
                        💬 Chat
                    </button>
                </div>
            </div>
        </div>

        <div v-else-if="query && !loading" class="empty-state">
            <span class="empty-emoji">🤷</span>
            <p class="empty-text">No users found</p>
            <p class="empty-hint">Try searching for a different username</p>
        </div>
    </div>
</template>

<style scoped>
.search-container {
    max-width: 700px;
    margin: 2rem auto;
    padding: 1.5rem;
}

.search-header {
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid #f0f0f0;
}

.search-title {
    font-size: 1.75rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
}

.search-box {
    display: flex;
    gap: 0.75rem;
    margin-bottom: 2rem;
    background: white;
    padding: 0.5rem;
    border-radius: 25px;
    box-shadow: var(--shadow-sm);
    transition: all 0.3s ease;
}

.search-box:focus-within {
    box-shadow: 0 6px 25px rgba(102, 126, 234, 0.15);
    transform: translateY(-2px);
}

.search-input {
    flex: 1;
    border: none;
    outline: none;
    padding: 0.75rem 1.25rem;
    font-size: 1rem;
    background: transparent;
    color: var(--text-primary);
}

.search-input::placeholder {
    color: #a0aec0;
}

/* spinner provided by shared.css */

.results-container {
    margin-top: 2rem;
}

.results-count {
    font-size: 0.9rem;
    color: var(--text-secondary);
    margin-bottom: 1rem;
    font-weight: 500;
}

.user-cards {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.user-card {
    background: white;
    border-radius: var(--radius-md);
    padding: 1.25rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    box-shadow: var(--shadow-sm);
    transition: all 0.3s ease;
}

.user-card:hover {
    transform: translateY(-3px);
    box-shadow: var(--shadow-md);
}

.user-info {
    display: flex;
    align-items: center;
    gap: 1rem;
}

/* ✅ ADD THESE STYLES */
.avatar-wrapper {
    flex-shrink: 0;
}

.user-avatar {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    object-fit: cover;
    border: 3px solid #e8f0fe;
}

.avatar-initials {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 1rem;
    border: 3px solid #e8f0fe;
}

.username {
    font-size: 1.05rem;
    font-weight: 600;
    color: var(--text-primary);
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

@media (max-width: 640px) {
    .search-container {
        padding: 1rem;
    }
    
    .search-box {
        flex-direction: column;
        border-radius: 20px;
    }
    
    .user-card {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
    }
    
    .user-info {
        flex-direction: column;
        text-align: center;
    }
}
</style>
