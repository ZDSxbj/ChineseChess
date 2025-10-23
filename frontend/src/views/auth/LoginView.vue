<script lang="ts" setup>
import type { CustomFormData } from '@/composables/FormExam'

import { ref, useTemplateRef } from 'vue'
import { login } from '@/api/user/login'
import FormContainer from '@/components/FormContainer.vue'
import { showMsg } from '@/components/MessageBox.tsx'

import { useFormExam } from '@/composables/FormExam'
import apiBus from '@/utils/apiBus'

const rememberMe = useTemplateRef('rememberMe')

const loginForm = ref<CustomFormData[]>([
  {
    id: 'email',
    value: '',
    label: '邮箱',
    type: 'email',
    autocomplete: 'email',
  },
  {
    id: 'password',
    value: '',
    label: '密码',
    type: 'password',
    autocomplete: 'current-password',
  },
])

const correct = useFormExam(loginForm)

async function loginAction() {
  const email = loginForm.value[0].value
  const password = loginForm.value[1].value

  try {
    const resp = await login({ email, password })

    apiBus.emit('API:LOGIN', resp)

    if (rememberMe.value?.checked) {
      localStorage.setItem('email', email)
      localStorage.setItem('password', password)
    }
  }
  catch (error) {
    console.error('Login failed:', error)
    showMsg(error as string)
  }
}
</script>

<template>
  <FormContainer
    class="mt-5 w-full sm:w-1/2"
    :form-data="loginForm"
    :disabled="!correct"
    @submit-form="loginAction"
  >
    <div class="flex flex-row justify-between">
      <div class="mx-a">
        <input ref="rememberMe" name="rememberMe" type="checkbox">
        <label for="rememberMe">记住我</label>
        <span class="text-sm text-gray-500">下次自动登录</span>
      </div>
      <RouterLink class="mx-a text-purple-800" to="/auth/register">
        注册个账号先
      </RouterLink>
    </div>
  </FormContainer>
</template>
