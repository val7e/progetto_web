<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const isLoginPage = computed(() => route.path === '/login')
const currentUsername = computed(() => localStorage.getItem('username') || '')
</script>

<script>
export default {}
</script>

<template>
    <!-- Header - hidden on login page -->
    <header v-if="!isLoginPage" class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
        <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WasaText</a>
        <button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
            <span class="navbar-toggler-icon"></span>
        </button>
    </header>

    <!-- Login page - full width -->
    <div v-if="isLoginPage" class="container-fluid">
        <RouterView />
    </div>

    <!-- Other pages - with sidebar -->
    <div v-else class="container-fluid">
        <div class="row">
            <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
                <div class="position-sticky pt-3 sidebar-sticky">
                    <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
                        <span>Menu</span>
                    </h6>
                    <ul class="nav flex-column">
                        <li class="nav-item">
                            <RouterLink to="/home" class="nav-link">
                                <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#message-square"/></svg>
                                Conversations
                            </RouterLink>
                        </li>
                        <li class="nav-item">
                            <RouterLink to="/profile" class="nav-link">
                                <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#user"/></svg>
                                Profile <span v-if="currentUsername">({{ currentUsername }})</span>
                            </RouterLink>
                        </li>
                        <li class="nav-item">
                            <RouterLink to="/search" class="nav-link">
                                <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
                                Search user
                            </RouterLink>
                        </li>
                    </ul>
                </div>
            </nav>

            <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
                <RouterView />
            </main>
        </div>
    </div>
</template>

<style>
</style>
