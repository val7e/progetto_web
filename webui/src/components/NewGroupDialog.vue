<script>
export default {
    props: {
        show: {
            type: Boolean,
            required: true
        }
    },
    emits: ['close', 'group-created'],
    data() {
        return {
            newGroupName: '',
            newGroupMembers: [],
            groupSelectedPhoto: null,
            groupPhotoPreview: null,
            groupSearchQuery: '',
            groupSearchLoading: false,
            groupSearchResults: [],
            creatingGroup: false,
            errormsg: null
        }
    },
    methods: {
        close() {
            this.$emit('close');
            this.resetForm();
        },
        resetForm() {
            this.newGroupName = '';
            this.newGroupMembers = [];
            this.groupSelectedPhoto = null;
            this.groupPhotoPreview = null;
            this.groupSearchQuery = '';
            this.groupSearchResults = [];
            this.errormsg = null;
            if (this.$refs.groupPhotoInput) {
                this.$refs.groupPhotoInput.value = '';
            }
        },
        removeMemberFromNewGroup(idx) {
            this.newGroupMembers.splice(idx, 1);
        },
        onGroupPhotoSelected(e) {
            const file = e.target.files[0];
            if (!file) return;
            if (!file.type.startsWith('image/')) {
                this.errormsg = 'Please select an image file';
                return;
            }
            this.groupSelectedPhoto = file;
            const reader = new FileReader();
            reader.onload = ev => this.groupPhotoPreview = ev.target.result;
            reader.readAsDataURL(file);
        },
        async searchGroupUsers() {
            const q = (this.groupSearchQuery || '').trim();
            this.groupSearchResults = [];
            if (!q) return;
            this.groupSearchLoading = true;
            try {
                const res = await this.$axios.get(`/users?searcheduser=${encodeURIComponent(q)}`);
                this.groupSearchResults = res.data || [];
            } catch (e) {
                console.error('User search failed:', e);
            }
            this.groupSearchLoading = false;
        },
        addMemberFromSearch(username) {
            if (!username) return;
            if (!this.newGroupMembers.includes(username)) {
                this.newGroupMembers.push(username);
            }
        },
        async createGroupConfirm() {
    if (!this.newGroupName.trim()) {
        this.errormsg = 'Group name is required';
        return;
    }
    this.creatingGroup = true;
    this.errormsg = null;
    try {
        console.log('Creating group with name:', this.newGroupName); // Debug
        
        // 1. Create group
        const res = await this.$axios.post('/groups', { 
            name: this.newGroupName.trim() 
        });
        const group = res.data;
        console.log('Group created successfully:', group); // Debug

        // 2. Add members if any
        if (this.newGroupMembers.length > 0) {
            console.log('Adding members:', this.newGroupMembers); // Debug
            await this.$axios.post(`/groups/${group.id}/members`, { 
                members: this.newGroupMembers 
            });
            console.log('Members added successfully'); // Debug
        }

        // 3. Upload photo if present
        if (this.groupPhotoPreview) {
            console.log('Uploading group photo'); // Debug
            const base64 = this.groupPhotoPreview.split(',')[1];
            await this.$axios.put(`/groups/${group.id}/photo`, { 
                photo: base64 
            });
            console.log('Photo uploaded successfully'); // Debug
        }

        console.log('Emitting group-created event with ID:', group.id); // Debug
        this.$emit('group-created', group.id);
        this.close();
    } catch (e) {
        console.error('Error creating group:', e); // Debug
        console.error('Error response:', e.response?.data); // Debug
        
        // Better error message
        if (e.response?.data?.error) {
            this.errormsg = e.response.data.error;
        } else if (e.response?.data?.message) {
            this.errormsg = e.response.data.message;
        } else {
            this.errormsg = e.message || e.toString();
        }
    }
    this.creatingGroup = false;
}
    }
}
</script>

<template>
    <div v-if="show" class="modal-overlay" @click="close">
        <div class="modal-dialog modal-large" @click.stop>
            <div class="modal-header">
                <h3>👥 Create New Group</h3>
                <button class="btn-modal-close" @click="close">✕</button>
            </div>
            <div class="modal-body">
                <ErrorMsg v-if="errormsg" :msg="errormsg" />
                
                <!-- Group Name -->
                <div class="settings-section">
                    <label class="modal-label">Group Name *</label>
                    <input 
                        v-model="newGroupName" 
                        type="text" 
                        class="input-cute" 
                        placeholder="Enter group name" 
                        style="width: 100%;"
                    />
                </div>

                <!-- Group Photo -->
                <div class="settings-section">
                    <label class="modal-label">Group Photo (Optional)</label>
                    <div class="photo-upload-section">
                        <input 
                            ref="groupPhotoInput"
                            type="file" 
                            class="file-input-hidden" 
                            accept="image/*" 
                            @change="onGroupPhotoSelected" 
                            id="groupPhotoFile" 
                        />
                        <label for="groupPhotoFile" class="btn-cute btn-secondary">
                            📷 Choose Photo
                        </label>
                    </div>
                    <div v-if="groupPhotoPreview" class="photo-preview-small">
                        <img :src="groupPhotoPreview" alt="Preview" />
                    </div>
                </div>

                <!-- Current Members -->
                <div class="settings-section">
                    <label class="modal-label">Members ({{ newGroupMembers.length }})</label>
                    <div class="members-chips">
                        <span v-for="(m, idx) in newGroupMembers" :key="m" class="member-chip">
                            {{ m }}
                            <button class="chip-remove" @click="removeMemberFromNewGroup(idx)">✕</button>
                        </span>
                        <div v-if="newGroupMembers.length === 0" class="no-members">
                            No members added yet
                        </div>
                    </div>
                </div>

                <!-- Search & Add Members -->
                <div class="settings-section">
                    <label class="modal-label">Add Members</label>
                    <div class="search-group">
                        <input 
                            v-model="groupSearchQuery" 
                            @keyup.enter="searchGroupUsers" 
                            type="text" 
                            class="modal-input" 
                            placeholder="Search username..." 
                        />
                        <button class="btn-modal-search" @click="searchGroupUsers" :disabled="groupSearchLoading">
                            🔍
                        </button>
                    </div>
                    
                    <div v-if="groupSearchResults && groupSearchResults.length" class="search-results">
                        <div v-for="u in groupSearchResults" :key="u.username" class="result-item">
                            <div class="result-user">
                                <img v-if="u.pic" :src="`data:image/png;base64,${u.pic}`" class="result-avatar" />
                                <div v-else class="result-avatar-placeholder">{{ u.username[0].toUpperCase() }}</div>
                                <span>{{ u.username }}</span>
                            </div>
                            <button 
                                class="btn-result-action" 
                                @click="addMemberFromSearch(u.username)"
                                :disabled="newGroupMembers.includes(u.username)"
                            >
                                {{ newGroupMembers.includes(u.username) ? '✓ Added' : '➕ Add' }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-modal-cancel" @click="close">Cancel</button>
                <button 
                    class="btn-cute btn-success" 
                    :disabled="creatingGroup || !newGroupName.trim()" 
                    @click="createGroupConfirm"
                >
                    <span v-if="creatingGroup" class="spinner-sm"></span>
                    {{ creatingGroup ? 'Creating...' : '✨ Create Group' }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.members-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    min-height: 2rem;
}

.member-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: var(--bg-light);
    color: var(--text-primary);
    border: 2px solid var(--border-color);
    border-radius: 16px;
    padding: 0.25rem 0.5rem;
    font-size: 0.9rem;
    font-weight: 500;
}

.chip-remove {
    background: var(--danger-gradient);
    color: white;
    border: none;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    line-height: 1;
}

.no-members {
    color: var(--text-secondary);
    font-style: italic;
}
</style>
