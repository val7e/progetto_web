<script>
export default {
    data: function() {
        return {
            errormsg: null,
            loading: false,
            profile: null,
            newUsername: '',
            editingUsername: false,
            selectedPhoto: null,
            photoPreview: null
        }
    },
    methods: {
        async refresh() {
            this.loading = true;
            this.errormsg = null;
            try {
                const current = localStorage.getItem('username')
                let response = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(current)}`);
                const me = (response.data || []).find(u => u.username === current) || null
                this.profile = me;
                this.newUsername = me?.username || '';
            } catch (e) {
                this.errormsg = e.toString();
            }
            this.loading = false;
        },
        
        async updateUsername() {
            if (!this.newUsername.trim()) return;
            
            this.loading = true;
            this.errormsg = null;
            try {
                await this.$axios.put("/users/me/username", {
                    username: this.newUsername
                });
                localStorage.setItem('username', this.newUsername);
                this.editingUsername = false;
                await this.refresh();
            } catch (e) {
                this.errormsg = e.toString();
            }
            this.loading = false;
        },
        onPhotoSelected(e) {
            const file = e.target.files[0]
            if (!file) return
            if (!file.type.startsWith('image/')) {
                this.errormsg = 'Please select an image file'
                return
            }
            const reader = new FileReader()
            reader.onload = (ev) => { this.photoPreview = ev.target.result }
            reader.readAsDataURL(file)
            this.selectedPhoto = file
        },
        async uploadPhoto() {
            if (!this.photoPreview) return
            this.loading = true
            this.errormsg = null
            try {
                const base64 = this.photoPreview.split(',')[1]
                await this.$axios.put('/users/me/pic', { pic: base64 })
                await this.refresh()
                this.selectedPhoto = null
                this.photoPreview = null
                this.$refs.photoInput.value = ''
            } catch (e) {
                this.errormsg = e.toString()
            }
            this.loading = false
        },
        
        logout() {
            localStorage.clear();
            this.$router.push('/login');
        }
    },
    mounted() {
        this.refresh()
    }
}
</script>

<template>
    <div class="profile-container">
        <div class="profile-header">
            <h1 class="profile-title">✨ My Profile</h1>
            <button type="button" class="btn-cute btn-danger" @click="logout">
                👋 Logout
            </button>
        </div>

        <ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

        <div v-if="loading" class="text-center my-3">
            <div class="spinner-border text-primary" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
        </div>

        <div v-else-if="profile" class="profile-card">
            <!-- Profile Picture Section -->
            <div class="profile-pic-section">
                <div class="pic-wrapper">
                    <img 
                        v-if="profile.pic" 
                        :src="`data:image/png;base64,${profile.pic}`" 
                        alt="Profile" 
                        class="profile-avatar"
                    />
                    <div v-else class="avatar-placeholder" style="width: 150px; height: 150px; font-size: 4rem;">
                        <span>👤</span>
                    </div>
                </div>
                
                <div class="pic-upload">
                    <input 
                        ref="photoInput" 
                        type="file" 
                        class="file-input-hidden" 
                        accept="image/*" 
                        @change="onPhotoSelected" 
                        id="photoUpload"
                    />
                    <label for="photoUpload" class="btn-cute btn-primary">
                        📷 Choose Photo
                    </label>
                    <button 
                        class="btn-cute btn-secondary" 
                        :disabled="!photoPreview" 
                        @click="uploadPhoto"
                    >
                        ⬆️ Upload
                    </button>
                </div>

                <div v-if="photoPreview" class="photo-preview">
                    <p class="preview-label">Preview:</p>
                    <img :src="photoPreview" alt="Preview" class="preview-img" />
                </div>
            </div>

            <!-- Username Section -->
            <div class="username-section">
                <label class="section-label">💬 Username</label>
                
                <div v-if="!editingUsername" class="username-display">
                    <span class="username-text">{{ profile.username }}</span>
                    <button @click="editingUsername = true" class="btn-cute btn-primary" style="padding: 0.5rem 1rem; font-size: 0.85rem;">
                        ✏️ Edit
                    </button>
                </div>
                
                <div v-else class="username-edit">
                    <input 
                        v-model="newUsername" 
                        type="text" 
                        class="input-cute" 
                        placeholder="Enter new username"
                    />
                    <div class="edit-actions">
                        <button @click="updateUsername" class="btn-cute btn-success">💾 Save</button>
                        <button @click="editingUsername = false" class="btn-cute" style="background: var(--bg-light); color: var(--text-secondary); border: 2px solid var(--border-color);">❌ Cancel</button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>


<style scoped>
.profile-container {
    max-width: 600px;
    margin: 2rem auto;
    padding: 1.5rem;
}

.profile-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid #f0f0f0;
}

.profile-title {
    font-size: 1.75rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
}

.profile-card {
    background: white;
    border-radius: var(--radius-lg);
    padding: 2rem;
    box-shadow: var(--shadow-sm);
}

.profile-pic-section {
    text-align: center;
    padding-bottom: 2rem;
    border-bottom: 1px solid #f0f0f0;
    margin-bottom: 2rem;
}

.pic-wrapper {
    margin-bottom: 1.5rem;
}

.profile-avatar {
    width: 150px;
    height: 150px;
    border-radius: 50%;
    object-fit: cover;
    border: 4px solid #e8f0fe;
    box-shadow: var(--shadow-sm);
}

.pic-upload {
    display: flex;
    gap: 0.75rem;
    justify-content: center;
    align-items: center;
    flex-wrap: wrap;
}

.photo-preview {
    margin-top: 1.5rem;
}

.preview-label {
    font-size: 0.85rem;
    color: var(--text-secondary);
    margin-bottom: 0.5rem;
}

.preview-img {
    width: 120px;
    height: 120px;
    border-radius: 15px;
    object-fit: cover;
    border: 3px solid #e8f0fe;
    box-shadow: var(--shadow-sm);
}

.username-section {
    margin-top: 1.5rem;
}

.section-label {
    display: block;
    font-size: 0.95rem;
    font-weight: 600;
    color: #4a5568;
    margin-bottom: 0.75rem;
}

.username-display {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: var(--bg-light);
    border-radius: var(--radius-sm);
}

.username-text {
    flex: 1;
    font-size: 1.1rem;
    font-weight: 500;
    color: var(--text-primary);
}

.username-edit {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.edit-actions {
    display: flex;
    gap: 0.5rem;
}

@media (max-width: 640px) {
    .profile-container {
        padding: 1rem;
    }
    
    .profile-header {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
    }
    
    .profile-card {
        padding: 1.5rem;
    }
}
</style>
