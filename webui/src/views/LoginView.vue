<script>
export default {
    data: function() {
        return {
            errormsg: null,
            loading: false,
            username: ''
        }
    },
    methods: {
        async handleLogin() {
            if (!this.username.trim()) {
                this.errormsg = "Please enter a username";
                return;
            }

            this.loading = true;
            this.errormsg = null;
            try {
                let response = await this.$axios.post("/session", { 
                    username: this.username 
                });
                console.log('Login response:', response.data);  
                localStorage.setItem('userToken', response.data.identifier);
                localStorage.setItem('username', response.data.username);
                
                this.$router.push('/home');
            } catch (e) {
                this.errormsg = e.response?.data?.error || e.toString();
            }
            this.loading = false;
        }
    }
}
</script>

<template>
    <div class="login-page">
        <div class="login-overlay"></div>
        <div class="login-container">
            <div class="login-card">
                <div class="login-header">
                    <h1 class="app-title">
                        <span class="title-emoji">📬</span>
                        <span class="title-text">WasaText</span>
                        <span class="title-emoji">🤖</span>
                    </h1>
                    <p class="app-subtitle">Connect with your friends effortlessly!</p>
                </div>
                
                <ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
                
                <div class="login-form">
                    <div class="form-group">
                        <label for="username" class="form-label">Username</label>
                        <input 
                            id="username"
                            v-model="username" 
                            type="text" 
                            class="input-cute" 
                            placeholder="Enter your username"
                            @keyup.enter="handleLogin"
                            :disabled="loading"
                        />
                    </div>
                    
                    <button 
                        @click="handleLogin" 
                        class="btn-cute btn-primary btn-login"
                        :disabled="loading"
                    >
                        <span v-if="loading" class="spinner-sm"></span>
                        {{ loading ? 'Logging in...' : '🚀 Login' }}
                    </button>
                </div>
                
                <div class="login-footer">
                    <p class="footer-text">Welcome!👋</p>
                    <p class="footer-text">If you're new, enter a username to register and click Login!</p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.login-page {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
    background-attachment: fixed;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
}

.login-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.3) 0%, rgba(118, 75, 162, 0.3) 100%);
    backdrop-filter: blur(2px);
}

.login-container {
    position: relative;
    z-index: 10;
    width: 100%;
    max-width: 450px;
    padding: 2rem;
    animation: slideUp 0.6s ease-out;
}

@keyframes slideUp {
    from {
        opacity: 0;
        transform: translateY(30px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.login-card {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 24px;
    padding: 3rem 2.5rem;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    border: 1px solid rgba(255, 255, 255, 0.5);
}

.login-header {
    text-align: center;
    margin-bottom: 2.5rem;
}

.app-title {
    font-size: 2.5rem;
    font-weight: 700;
    margin: 0 0 0.5rem 0;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
}

.title-text {
    background: var(--primary-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    letter-spacing: -1px;
}

.title-emoji {
    font-size: 2rem;
    animation: bounce 2s infinite;
}

.title-emoji:first-child {
    animation-delay: 0s;
}

.title-emoji:last-child {
    animation-delay: 0.3s;
}

@keyframes bounce {
    0%, 100% {
        transform: translateY(0);
    }
    50% {
        transform: translateY(-10px);
    }
}

.app-subtitle {
    font-size: 1rem;
    color: var(--text-secondary);
    margin: 0;
    font-weight: 500;
}

.login-form {
    margin-bottom: 1.5rem;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-label {
    display: block;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
    font-size: 0.95rem;
}

.btn-login {
    width: 100%;
    padding: 0.9rem 1.5rem;
    font-size: 1.05rem;
    font-weight: 600;
    margin-top: 0.5rem;
}

.login-footer {
    text-align: center;
    padding-top: 1.5rem;
    border-top: 1px solid rgba(0, 0, 0, 0.1);
}

.footer-text {
    color: var(--text-secondary);
    margin: 0;
    font-size: 0.95rem;
}

/* spinner-sm now provided by shared.css */

@media (max-width: 640px) {
    .login-container {
        padding: 1rem;
    }
    
    .login-card {
        padding: 2rem 1.5rem;
    }
    
    .app-title {
        font-size: 2rem;
    }
    
    .title-emoji {
        font-size: 1.5rem;
    }
}
</style>
