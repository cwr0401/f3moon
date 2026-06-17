<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const isRegister = ref(false)
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const nickname = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)

function switchMode(register: boolean) {
  isRegister.value = register
  confirmPassword.value = ''
  error.value = ''
  success.value = ''
}

async function handleSubmit() {
  error.value = ''
  success.value = ''

  if (isRegister.value && password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true

  try {
    if (isRegister.value) {
      await auth.register(email.value, password.value, nickname.value)
      success.value = '注册成功，验证邮件已发送，请查收邮箱'
      password.value = ''
      confirmPassword.value = ''
    } else {
      await auth.login(email.value, password.value)
      router.push('/lobby')
    }
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    error.value = err.response?.data?.error || '操作失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-bg-dark px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-accent mb-2">荆楚花牌</h1>
        <p class="text-gray-400">花好月圆</p>
      </div>

      <div class="bg-bg-card rounded-xl p-8 border border-amber-900/30 shadow-2xl">
        <div class="flex mb-6 border-b border-amber-900/30">
          <button
            @click="switchMode(false)"
            :class="['flex-1 pb-3 text-center transition-colors', !isRegister ? 'text-accent border-b-2 border-accent' : 'text-gray-400 hover:text-gray-200']"
          >
            登录
          </button>
          <button
            @click="switchMode(true)"
            :class="['flex-1 pb-3 text-center transition-colors', isRegister ? 'text-accent border-b-2 border-accent' : 'text-gray-400 hover:text-gray-200']"
          >
            注册
          </button>
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label class="block text-sm text-gray-400 mb-1">邮箱</label>
            <input
              v-model="email"
              type="email"
              required
              class="w-full px-4 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/50"
              placeholder="your@email.com"
            />
          </div>

          <div v-if="isRegister">
            <label class="block text-sm text-gray-400 mb-1">昵称</label>
            <input
              v-model="nickname"
              type="text"
              class="w-full px-4 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/50"
              placeholder="牌桌上的名字"
            />
          </div>

          <div>
            <label class="block text-sm text-gray-400 mb-1">密码</label>
            <input
              v-model="password"
              type="password"
              required
              class="w-full px-4 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/50"
              placeholder="••••••••"
            />
          </div>

          <div v-if="isRegister">
            <label class="block text-sm text-gray-400 mb-1">确认密码</label>
            <input
              v-model="confirmPassword"
              type="password"
              required
              class="w-full px-4 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/50"
              placeholder="再次输入密码"
            />
          </div>

          <div v-if="error" class="text-red-400 text-sm bg-red-900/20 px-3 py-2 rounded">
            {{ error }}
          </div>

          <div v-if="success" class="text-green-400 text-sm bg-green-900/20 px-3 py-2 rounded">
            {{ success }}
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors disabled:opacity-50"
          >
            {{ loading ? '处理中...' : (isRegister ? '注册' : '登录') }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
