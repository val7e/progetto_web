import {createRouter, createWebHashHistory} from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import ChatView from '../views/ChatView.vue'
import ProfileView from '../views/ProfileView.vue'
import SearchView from '../views/SearchView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/', 
			redirect: '/login'
		},
		{
			path: '/login',
			component: LoginView
		},
		{
			path: '/home',
			name: 'home',
			component: HomeView
		},
		{
			path: '/chat/:conversationId',
			name: 'chat',
			component: ChatView
		},
		{
			path: '/profile',
			name: 'profile',
			component: ProfileView
		},
		{
			path: '/search',
			name: 'search',
			component: SearchView
        }
        }
    ]
})

// Simple auth guard based on presence of a stored token
router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('userToken')
    if (to.path !== '/login' && !token) {
        next('/login')
        return
    }
    if (to.path === '/login' && token) {
        next('/home')
        return
    }
    next()
})

export default router
