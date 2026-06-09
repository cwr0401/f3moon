import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const LoginView = () => import('../views/LoginView.vue')
const LobbyView = () => import('../views/LobbyView.vue')
const RoomView = () => import('../views/RoomView.vue')
const GameView = () => import('../views/GameView.vue')

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { guest: true } },
    { path: '/lobby', name: 'lobby', component: LobbyView },
    { path: '/room/:id', name: 'room', component: RoomView },
    { path: '/game/:id', name: 'game', component: GameView },
    { path: '/', redirect: '/lobby' },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!auth.isLoggedIn && to.name !== 'login') return { name: 'login' }
  if (auth.isLoggedIn && to.meta.guest) return { name: 'lobby' }
})

export default router
