import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import api from '../composables/useApi'

export interface User {
  id: string
  email: string
  nickname: string
  verified: boolean
  created_at: string
  updated_at: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const user = ref<User | null>(null)
  const isLoggedIn = computed(() => !!token.value)

  function init() {
    const saved = localStorage.getItem('token')
    if (saved) {
      token.value = saved
      fetchMe()
    }
  }

  async function login(email: string, password: string) {
    const { data } = await api.post('/auth/login', { email, password })
    token.value = data.token
    user.value = data.user
    localStorage.setItem('token', data.token)
  }

  async function register(email: string, password: string, nickname: string) {
    await api.post('/auth/register', { email, password, nickname })
  }

  async function fetchMe() {
    try {
      const { data } = await api.get('/auth/me')
      user.value = data
    } catch {
      logout()
    }
  }

  async function resendVerification(email: string) {
    await api.post('/auth/resend', { email })
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  return { token, user, isLoggedIn, init, login, register, fetchMe, resendVerification, logout }
})
