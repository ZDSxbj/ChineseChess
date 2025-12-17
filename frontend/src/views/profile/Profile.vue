<template>
  <div class="max-w-4xl mx-auto p-8 animate-fade-in">
    <!-- 加载状态 -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center py-20 text-amber-800/60">
      <div class="w-12 h-12 border-4 border-amber-200 border-t-amber-600 rounded-full animate-spin mb-4"></div>
      <p class="font-medium">正在加载数据...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="hasError" class="text-center py-20">
      <p class="text-red-600 mb-4 font-medium">加载失败，请稍后重试。</p>
      <button @click="initForm()" class="px-6 py-2 bg-amber-100 text-amber-900 rounded-lg hover:bg-amber-200 transition-colors font-bold">重试</button>
    </div>

    <!-- 表单内容区 -->
    <form v-else class="relative bg-white/60 backdrop-blur-sm rounded-2xl shadow-sm border border-amber-100 p-8 overflow-hidden" @submit.prevent>
      <!-- 装饰元素 -->
      <div class="absolute top-0 left-0 w-16 h-16 border-t-4 border-l-4 border-amber-200/60 rounded-tl-2xl pointer-events-none"></div>
      <div class="absolute bottom-0 right-0 w-16 h-16 border-b-4 border-r-4 border-amber-200/60 rounded-br-2xl pointer-events-none"></div>
      <div class="absolute top-4 right-4 text-amber-900/5 pointer-events-none">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-32 h-32">
          <path fill-rule="evenodd" d="M7.5 6a4.5 4.5 0 1 1 9 0 4.5 4.5 0 0 1-9 0ZM3.751 20.105a8.25 8.25 0 0 1 16.498 0 .75.75 0 0 1-.437.695A18.683 18.683 0 0 1 12 22.5c-2.786 0-5.433-.608-7.812-1.7a.75.75 0 0 1-.437-.695Z" clip-rule="evenodd" />
        </svg>
      </div>

      <div class="flex items-center py-6 border-b border-amber-200/50 relative z-10">
        <label class="w-24 font-bold text-amber-900">头像</label>
        <div class="flex items-center gap-6">
          <!-- 如果没有头像则不渲染 img，保持空白 -->
          <div class="relative group">
            <img v-if="formData.avatar" :src="formData.avatar" alt="" class="w-24 h-24 rounded-full object-cover border-4 border-amber-200 shadow-md transition-transform group-hover:scale-105" @error="handleAvatarError" />
            <div v-else class="w-24 h-24 rounded-full bg-amber-100 border-4 border-amber-200 shadow-md flex items-center justify-center text-amber-300">
              <div class="i-carbon-user-avatar text-4xl"></div>
            </div>
          </div>

          <input type="file" id="avatar-input" class="hidden" accept="image/*" @change="handleAvatarChange" :disabled="isUploadingAvatar" />
          <label for="avatar-input" class="px-4 py-2 bg-white border border-amber-300 text-amber-800 rounded-lg cursor-pointer hover:bg-amber-50 transition-colors font-bold shadow-sm text-sm flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
            </svg>
            {{ isDefaultAvatar ? '上传头像' : '修改头像' }}
          </label>
        </div>
      </div>

      <!-- 姓名：显示 + 箭头编辑 -->
      <div class="flex items-center py-6 border-b border-amber-200/50 relative z-10">
        <label class="w-24 font-bold text-amber-900">姓名</label>
        <div class="flex-1">
          <div class="flex items-center justify-between group cursor-pointer" @click="startEditing('name')" v-if="editingField !== 'name'">
            <div class="text-amber-900 font-medium text-lg">{{ formData.name }}</div>
            <button type="button" class="text-amber-400 group-hover:text-amber-600 transition-colors p-2 hover:bg-amber-50 rounded-full border-none ring-0 outline-none">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
              </svg>
            </button>
          </div>
          <div v-else class="flex items-center gap-3 animate-fade-in">
            <input type="text" class="flex-1 px-4 py-2 border border-amber-300 rounded-lg focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none bg-white/80 text-amber-900" v-model="tempValue" placeholder="请输入姓名" />
            <div class="flex gap-2">
              <button type="button" class="px-3 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors shadow-sm flex items-center justify-center border-none ring-0 outline-none" @click="saveEdit('name')">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                </svg>
              </button>
              <button type="button" class="px-3 py-2 bg-gray-200 text-gray-600 rounded-lg hover:bg-gray-300 transition-colors flex items-center justify-center border-none ring-0 outline-none" @click="cancelEdit()">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 性别 -->
      <div class="flex items-center py-6 border-b border-amber-200/50 relative z-10">
        <label class="w-24 font-bold text-amber-900">性别</label>
        <div class="flex-1">
          <div class="flex items-center justify-between group cursor-pointer" @click="startEditing('gender')" v-if="editingField !== 'gender'">
            <div class="text-amber-900 font-medium text-lg">{{ formData.gender }}</div>
            <button type="button" class="text-amber-400 group-hover:text-amber-600 transition-colors p-2 hover:bg-amber-50 rounded-full border-none ring-0 outline-none">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
              </svg>
            </button>
          </div>
          <div v-else class="flex items-center gap-3 animate-fade-in">
            <select class="flex-1 px-4 py-2 border border-amber-300 rounded-lg focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none bg-white/80 text-amber-900" v-model="tempValue">
              <option value="">请选择性别</option>
              <option value="男">男</option>
              <option value="女">女</option>
              <option value="其他">其他</option>
            </select>
            <div class="flex gap-2">
              <button type="button" class="px-3 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors shadow-sm flex items-center justify-center border-none ring-0 outline-none" @click="saveEdit('gender')">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                </svg>
              </button>
              <button type="button" class="px-3 py-2 bg-gray-200 text-gray-600 rounded-lg hover:bg-gray-300 transition-colors flex items-center justify-center border-none ring-0 outline-none" @click="cancelEdit()">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 邮箱 -->
      <div class="flex items-center py-6 border-b border-amber-200/50 relative z-10">
        <label class="w-24 font-bold text-amber-900">邮箱</label>
        <div class="flex-1">
          <div class="flex items-center justify-between group cursor-pointer" @click="startEditing('email')" v-if="editingField !== 'email'">
            <div class="text-amber-900 font-medium text-lg">{{ formData.email }}</div>
            <button type="button" class="text-amber-400 group-hover:text-amber-600 transition-colors p-2 hover:bg-amber-50 rounded-full border-none ring-0 outline-none">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
              </svg>
            </button>
          </div>
          <div v-else class="flex items-center gap-3 animate-fade-in">
            <input type="email" class="flex-1 px-4 py-2 border border-amber-300 rounded-lg focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none bg-white/80 text-amber-900" v-model="newEmail" placeholder="请输入邮箱" />
            <div class="flex gap-2">
              <button type="button" class="px-4 py-2 bg-amber-600 text-white rounded-lg hover:bg-amber-700 transition-colors shadow-sm text-sm font-bold whitespace-nowrap border-none ring-0 outline-none" :disabled="isSendingCode" @click="sendCode">
                <span v-if="!isSendingCode">发送验证码</span>
                <span v-else>{{ sendCountdown > 0 ? sendCountdown + 's' : '发送中...' }}</span>
              </button>
              <button type="button" class="px-3 py-2 bg-gray-200 text-gray-600 rounded-lg hover:bg-gray-300 transition-colors flex items-center justify-center border-none ring-0 outline-none" @click="cancelEdit()">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

    <div v-if="showEmailVerifyModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-[#fdf6e3] p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <h3 class="text-xl font-black text-amber-900 mb-4">验证邮箱</h3>
        <p class="text-amber-800/70 mb-4 text-sm">请输入收到的验证码，已发送到：<br/><span class="font-bold text-amber-900">{{ newEmail }}</span></p>
        <input type="text" class="w-full px-4 py-2 border border-amber-300 rounded-lg focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none bg-white text-amber-900 text-center tracking-widest text-lg font-bold mb-6" v-model="verifyCodeInput" placeholder="验证码" />
        <div class="flex gap-3 justify-center">
          <button class="px-6 py-2 rounded-lg border-2 border-amber-200 text-amber-800 hover:bg-amber-100 font-bold transition-colors" @click="closeVerifyModal">取消</button>
          <button class="px-6 py-2 rounded-lg bg-amber-600 text-white hover:bg-amber-700 font-bold shadow-md" :disabled="isVerifying" @click="confirmVerify">
            <span v-if="!isVerifying">确认</span>
            <span v-else>验证中...</span>
          </button>
        </div>
        <p v-if="verifyError" class="text-red-600 mt-4 text-sm font-medium">{{ verifyError }}</p>
      </div>
    </div>

    <div v-if="showEmailError" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-[#fdf6e3] p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <div class="text-amber-600 text-4xl mb-4 flex justify-center"><div class="i-carbon-warning-alt"></div></div>
        <p class="text-amber-900 font-bold mb-6">{{ emailErrorMsg }}</p>
        <div class="flex justify-center">
          <button class="px-6 py-2 rounded-lg bg-amber-600 text-white hover:bg-amber-700 font-bold shadow-md" @click="showEmailError = false">确定</button>
        </div>
      </div>
    </div>

      <!-- 经验值（只读，不可编辑） -->
      <div class="flex items-center py-6 border-b border-amber-200/50">
        <label class="w-24 font-bold text-amber-900">经验值</label>
        <div class="flex-1 text-amber-900 font-medium text-lg">{{ formData.exp }}</div>
      </div>

      <!-- 总场次（只读，不可编辑） -->
      <div class="flex items-center py-6 border-b border-amber-200/50">
        <label class="w-24 font-bold text-amber-900">总场次</label>
        <div class="flex-1 text-amber-900 font-medium text-lg">{{ formData.totalGames }}</div>
      </div>

      <!-- 胜率（只读，不可编辑） -->
      <div class="flex items-center py-6 border-b border-amber-200/50">
        <label class="w-24 font-bold text-amber-900">胜率</label>
        <div class="flex-1 text-amber-900 font-medium text-lg">{{ (formData.winRate || 0).toFixed(2) }}%</div>
      </div>

      <div class="flex justify-center gap-4 mt-10">
           <button type="button" class="px-6 py-2.5 bg-white border border-amber-300 text-amber-800 rounded-lg hover:bg-amber-50 transition-colors font-bold shadow-sm" @click="openChangePwd">修改密码</button>
           <button type="button" class="px-6 py-2.5 bg-white border border-amber-300 text-amber-800 rounded-lg hover:bg-amber-50 transition-colors font-bold shadow-sm" @click="promptLogout">退出登录</button>
        <button type="button" class="px-6 py-2.5 bg-red-50 border border-red-200 text-red-600 rounded-lg hover:bg-red-100 transition-colors font-bold shadow-sm" @click="deleteAccount" :disabled="isProcessing">
          <span v-if="!isProcessing">注销账号</span>
          <span v-else>处理中...</span>
        </button>
      </div>
    </form>

    <!-- 页面内退出确认弹窗 -->
      <div v-if="showLogoutConfirm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
       <div class="bg-[#fdf6e3] p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
          <h3 class="text-xl font-black text-amber-900 mb-4">退出登录</h3>
          <p class="text-amber-800/80 mb-6 font-medium">确定要退出登录吗？</p>
          <div class="flex justify-center gap-3">
            <button class="px-6 py-2 rounded-lg border-2 border-amber-200 text-amber-800 hover:bg-amber-100 font-bold transition-colors" @click="cancelLogout">取消</button>
            <button class="px-6 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700 font-bold shadow-md" @click="logout">确定</button>
         </div>
        </div>
      </div>
    <!-- 保存成功提示弹窗 -->
      <div v-if="showSaveSuccess" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-[#fdf6e3] p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <div class="text-emerald-600 text-4xl mb-4 flex justify-center"><div class="i-carbon-checkmark-outline"></div></div>
        <p class="text-amber-900 font-bold mb-6 text-lg">保存成功</p>
        <div class="flex justify-center">
          <button class="px-6 py-2 rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 font-bold shadow-md" @click="closeSaveModal">确定</button>
        </div>
      </div>
    </div>

      <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-[#fdf6e3] p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <div class="text-red-600 text-4xl mb-4 flex justify-center"><div class="i-carbon-warning"></div></div>
        <h3 class="text-xl font-black text-amber-900 mb-4">注销账号</h3>
        <p class="text-amber-800/80 mb-6 font-medium text-sm">确定要注销账号吗？注销后所有个人数据将被删除，且无法恢复！</p>
        <div class="flex justify-center gap-3">
          <button class="px-6 py-2 rounded-lg border-2 border-amber-200 text-amber-800 hover:bg-amber-100 font-bold transition-colors" @click="showDeleteConfirm = false">取消</button>
          <button class="px-6 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700 font-bold shadow-md" @click="confirmDeleteAccount" :disabled="isProcessing">
            <span v-if="!isProcessing">确定</span>
            <span v-else>处理中...</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 修改密码弹窗 -->
    <div v-if="showChangePwdModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-[#fdf6e3] p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <h3 class="text-xl font-black text-amber-900 mb-4">修改密码</h3>
        <p class="text-amber-800/70 mb-4 font-medium" v-if="changePwdStep === 'old'">请输入原密码</p>
        <p class="text-amber-800/70 mb-4 font-medium" v-else>请输入新密码</p>

        <input v-if="changePwdStep === 'old'" type="password" class="w-full px-4 py-2 border border-amber-300 rounded-lg focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none bg-white text-amber-900 mb-4" v-model="oldPassword" placeholder="请输入原密码" />
        <input v-else type="password" class="w-full px-4 py-2 border border-amber-300 rounded-lg focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none bg-white text-amber-900 mb-2" v-model="newPassword" placeholder="请输入新密码" />

        <p v-if="changePwdStep === 'new'" class="text-xs text-amber-800/50 mb-4 text-left">
          密码至少6位，且包含大小写字母、数字和 !@#$%^&*? 中的一个。
        </p>

        <div class="flex justify-center gap-3">
          <button class="px-6 py-2 rounded-lg border-2 border-amber-200 text-amber-800 hover:bg-amber-100 font-bold transition-colors" @click="cancelChangePwd">取消</button>
          <button class="px-6 py-2 rounded-lg bg-amber-600 text-white hover:bg-amber-700 font-bold shadow-md" :disabled="isChangingPwd" @click="confirmChangePwd">
            <span v-if="!isChangingPwd">确定</span>
            <span v-else>处理中...</span>
          </button>
        </div>
        <p v-if="changePwdError" class="text-red-600 mt-4 text-sm font-medium">{{ changePwdError }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/store/useStore';
import RequestHandler, { API_URL } from '@/api/useRequest';
import { getProfile } from '@/api/user/getProfile';
import { updateProfile } from '@/api/user/updateProfile';
import { sendCode as sendEmailCode } from '@/api/user/send_code'
import { updateEmail } from '@/api/user/updateEmail'
import { updatePassword } from '@/api/user/updatePassword'
import { checkPassword } from '@/api/user/checkPassword'
import { uploadAvatar } from '@/api/user/uploadAvatar'
// 默认头像：后端存的是相对路径，前端拼接完整 URL
const DEFAULT_AVATAR_PATH = '/uploads/avatars/default.png'
const DEFAULT_AVATAR_URL = `${API_URL}${DEFAULT_AVATAR_PATH}`

// 路由实例
const router = useRouter();
// 用户 Store
const userStore = useUserStore();

// 状态管理
const isLoading = ref(true);
const hasError = ref(false);
const isProcessing = ref(false); // 处理状态（用于注销账号）
const isUploadingAvatar = ref(false);
const isDefaultAvatar = ref(true);

const showLogoutConfirm = ref(false);
const showSaveSuccess = ref(false);
const showEmailError = ref(false);
const emailErrorMsg = ref('');
const showDeleteConfirm = ref(false);
const closeSaveModal = () => { showSaveSuccess.value = false };


// 表单数据，添加总场次和胜率字段
const formData = reactive<any>({
  avatar: '',     // 如果未上传，保持空字符串（不显示 img）
  name: '',
  gender: '',
  email: '',
  exp: 0,
  totalGames: 0,  // 总场次
  winRate: 0      // 胜率（百分比）
});

// 编辑状态与临时值（避免直接修改 formData）
const editingField = ref<string | null>(null);
const tempValue = ref<any>(null);

const newEmail = ref<string>('');
const verifyCodeInput  = ref<string>('');
const showEmailVerifyModal = ref<boolean>(false);
const isSendingCode = ref<boolean>(false);
const sendCountdown = ref<number>(0);
let countdownTimer: number | null = null;
const isVerifying = ref<boolean>(false);
const verifyError = ref<string | null>(null);

// 修改密码弹窗相关状态
const showChangePwdModal = ref(false);
const oldPassword = ref('');
const newPassword = ref('');
const changePwdStep = ref<'old'|'new'>('old');
const changePwdError = ref<string | null>(null);
const isChangingPwd = ref(false);

/**
 * @description: 初始化表单数据（从 API 获取）
 */
const initForm = async () => {
  try {
    isLoading.value = true;
    hasError.value = false;

    // 保持原有调用逻辑；当前先不依赖后端结果（若后端可用会填充）
    const response = await getProfile().catch(() => null);
    const userData = (response && typeof response === 'object' && 'data' in response)
      ? (response as any).data
      : (response as any);

    if (userData) {
      Object.assign(formData, {
        // 头像规范化：若为相对路径则拼接 API_URL；为空则使用默认头像
        avatar: (() => {
          const a = userData.avatar || ''
          if (!a) return DEFAULT_AVATAR_URL
          if (typeof a === 'string' && a.startsWith('/uploads')) return `${API_URL}${a}`
          return a
        })(),
        name: userData.name || '未设置姓名',
        gender: userData.gender || '未选择',
        email: userData.email || '未设置邮箱',
        exp: userData.exp || 0,
        totalGames: userData.totalGames || 0,
        winRate: userData.winRate || 0
      });
      isDefaultAvatar.value = (userData.avatar == null) || (userData.avatar === DEFAULT_AVATAR_PATH) || (formData.avatar === DEFAULT_AVATAR_URL)
    } else {
      // 默认占位文本，avatar 保持空字符串
      Object.assign(formData, {
        avatar: DEFAULT_AVATAR_URL,
        name: '未设置姓名',
        gender: '未选择',
        email: '未设置邮箱',
        exp: 0,
        totalGames: 0,
        winRate: 0
      });
      isDefaultAvatar.value = true
    }

    isLoading.value = false;
  } catch (error) {
    console.error('获取个人资料失败:', error);
    hasError.value = true;
    isLoading.value = false;
  }
};

/* 编辑相关 */
const startEditing = (key: string) => {
  if (key === 'exp' || key === 'totalGames' || key === 'winRate') return;
  editingField.value = key;
  tempValue.value = formData[key];
  if (key === 'email') {
    newEmail.value = formData.email || '';
    verifyCodeInput.value = '';
    verifyError.value = null;
  }
};

const cancelEdit = () => {
  editingField.value = null;
  tempValue.value = null;
  newEmail.value = '';
  verifyCodeInput.value = '';
  showEmailVerifyModal.value = false;
  isSendingCode.value = false;
  sendCountdown.value = 0;
  verifyError.value = null;
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
};

// 打开修改密码弹窗
const openChangePwd = () => {
  showChangePwdModal.value = true;
  changePwdStep.value = 'old';
  oldPassword.value = '';
  newPassword.value = '';
  changePwdError.value = null;
};

// 取消修改密码
const cancelChangePwd = () => {
  showChangePwdModal.value = false;
  changePwdStep.value = 'old';
  oldPassword.value = '';
  newPassword.value = '';
  changePwdError.value = null;
  isChangingPwd.value = false;
};

// 确认修改密码
const confirmChangePwd = async () => {
  // 第一步：立即校验原密码
  if (changePwdStep.value === 'old') {
    if (!oldPassword.value) {
      changePwdError.value = '请输入原密码';
      return;
    }
    try {
      isChangingPwd.value = true;
      await checkPassword({ password: oldPassword.value });
      // 原密码正确，进入新密码输入
      changePwdStep.value = 'new';
      changePwdError.value = null;
    } catch (e: any) {
      const msg = e?.response?.data?.msg || '密码错误';
      changePwdError.value = msg;
      // 保持在输入原密码步骤
    } finally {
      isChangingPwd.value = false;
    }
    return;
  }

  // 第二步：提交新密码并修改
  if (changePwdStep.value === 'new') {
    if (!newPassword.value) {
      changePwdError.value = '请输入新密码';
      return;
    }
    // 新旧密码相同则提示未修改
    if (newPassword.value === oldPassword.value) {
      changePwdError.value = '密码未修改';
      return;
    }
    // 复杂度校验：至少6位，包含大小写字母、数字和特殊字符中的一个
    const hasMinLen = newPassword.value.length >= 6;
    const hasLower = /[a-z]/.test(newPassword.value);
    const hasUpper = /[A-Z]/.test(newPassword.value);
    const hasDigit = /\d/.test(newPassword.value);
    const hasSpecial = /[!@#$%^&*?]/.test(newPassword.value);
    if (!(hasMinLen && hasLower && hasUpper && hasDigit && hasSpecial)) {
      changePwdError.value = '修改密码失败';
      return;
    }
    try {
      isChangingPwd.value = true;
      const resp = await updatePassword({ oldPassword: oldPassword.value, newPassword: newPassword.value });
      const msg = (resp as any)?.data?.msg || '修改成功';
      if (msg !== '修改成功') {
        changePwdError.value = msg;
        // 如果后端仍返回密码错误（极端情况），回到第一步
        if (msg.includes('密码错误')) changePwdStep.value = 'old';
        return;
      }
      // 成功
      showChangePwdModal.value = false;
      showSaveSuccess.value = true;
    } catch (e: any) {
      const msg = e?.response?.data?.msg || '修改失败，请稍后重试';
      changePwdError.value = msg;
      if (msg.includes('密码错误')) changePwdStep.value = 'old';
    } finally {
      isChangingPwd.value = false;
    }
  }
};

const sendCode = async () => {
  const email = (newEmail.value || '').trim();
  const emailPattern = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
  if (!emailPattern.test(email)) {
    emailErrorMsg.value = '请输入有效邮箱';
    showEmailError.value = true;
    return;
  }

  // 与当前邮箱相同：直接提示并跳过发送验证码
  if ((formData.email || '').trim().toLowerCase() === email.toLowerCase()) {
    emailErrorMsg.value = '该邮箱已注册';
    showEmailError.value = true;
    return;
  }

  // 预检是否为其他用户已注册邮箱：
  // 利用后端 UpdateEmail 的“先查重再验码”逻辑，传入无效验证码触发查重。
  try {
    await updateEmail({ email, code: 'invalid' });
    // 理论上不会成功；若意外成功也不继续发送验证码
    emailErrorMsg.value = '该邮箱已注册';
    showEmailError.value = true;
    return;
  } catch (e: any) {
    const msg = e?.response?.data?.msg || '';
    if (msg === '该邮箱已存在') {
      emailErrorMsg.value = '该邮箱已注册';
      showEmailError.value = true;
      return;
    }
    // 其他错误（如验证码错误/过期等）则继续发送验证码
  }

  try {
    isSendingCode.value = true;
    await sendEmailCode({ email });
    showEmailVerifyModal.value = true;
    sendCountdown.value = 60;
    if (countdownTimer) clearInterval(countdownTimer);
    countdownTimer = window.setInterval(() => {
      sendCountdown.value -= 1;
      if (sendCountdown.value <= 0 && countdownTimer) {
        clearInterval(countdownTimer);
        countdownTimer = null;
        isSendingCode.value = false;
      }
    }, 1000);
  } catch (err) {
    emailErrorMsg.value = '验证码发送失败，请稍后重试';
    showEmailError.value = true;
    isSendingCode.value = false;
  }
};

// 关闭验证码弹窗（可返回重新输入邮箱）
const closeVerifyModal = () => {
  showEmailVerifyModal.value = false;
  verifyCodeInput.value = '';
  verifyError.value = null;
  // 允许重新发送
  isSendingCode.value = false;
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; sendCountdown.value = 0; }
};

// 确认验证码并更新邮箱：先调用后端验证接口 -> 验证通过再调用 updateProfile 更新
const confirmVerify = async () => {
  const code = (verifyCodeInput.value || '').trim();
  const email = (newEmail.value || '').trim();
  if (!code) {
    verifyError.value = '请输入验证码';
    return;
  }

  try {
    isVerifying.value = true;
    verifyError.value = null;

    // 一次性提交邮箱和验证码，由后端校验并修改
    const resp = await updateEmail({ email, code });
    const respData = (resp as any)?.data ?? resp;
    const ok = !respData?.msg || respData?.msg === '修改成功';

    if (!ok) {
      if (respData?.msg === '该邮箱已存在') {
        verifyError.value = '该邮箱已存在';
      } else {
        verifyError.value = respData?.msg || '验证码错误或已过期';
      }
      isVerifying.value = false;
      return;
    }

    // 成功后更新本地显示
    formData.email = email;
    // 同步更新全局 Store
    if (userStore.userInfo) {
      userStore.setUser({ ...userStore.userInfo, email: email });
    }
    editingField.value = null;
    showEmailVerifyModal.value = false;
    showSaveSuccess.value = true;

    // 清理
    newEmail.value = '';
    verifyCodeInput.value = '';
    isVerifying.value = false;
    if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; sendCountdown.value = 0; isSendingCode.value = false; }
  } catch (err) {
    console.error('确认验证码或更新邮箱失败:', err);
    if (err?.response?.data?.msg === '该邮箱已存在') {
      verifyError.value = '该邮箱已存在';
    } else {
      verifyError.value = '验证失败，请重试';
    }
    isVerifying.value = false;
  }
};

const saveEdit = async (key: string) => {
  // 如果是邮箱，走验证码流程：先把 tempValue 的值同步到 newEmail（兼容性）
  if (key === 'email') {
    newEmail.value = ((tempValue.value ?? newEmail.value) as string).trim();
    if (!newEmail.value) {
      emailErrorMsg.value = '请输入新的邮箱';
      showEmailError.value = true;
      return;
    }
    // 发验证码（sendCode 内会检查邮箱是否存在并弹窗处理）
    await sendCode();
    return;
  }

  // 简单校验示例（其他字段）
  if ((key === 'winRate' || key === 'exp' || key === 'totalGames') && (tempValue.value === '' || tempValue.value == null || isNaN(Number(tempValue.value)))) {
    alert('请输入有效数值');
    return;
  }
  try {
    await updateProfile({ [key]: tempValue.value });
    formData[key] = tempValue.value;
    // 同步更新全局 Store
    if (userStore.userInfo) {
      userStore.setUser({ ...userStore.userInfo, [key]: tempValue.value });
    }
    editingField.value = null;
    tempValue.value = null;
    showSaveSuccess.value = true;
  } catch (error: any) {
    // 判断是否为“该id已被注册”
    const msg = error?.response?.data?.msg || error?.message || '保存失败，请重试';
    emailErrorMsg.value = msg;
    showEmailError.value = true;
  }
};

/**
 * @description: 处理头像上传
 */
const handleAvatarError = () => {
  formData.avatar = DEFAULT_AVATAR_URL; // Fallback to default avatar
};
const handleAvatarChange = async (e: Event) => {
  const input = e.target as HTMLInputElement;
  if (!input.files || !input.files[0]) return;
  const file = input.files[0];
  try {
    isUploadingAvatar.value = true;
    // 调用后端上传接口，返回相对路径，如 /uploads/avatars/xxx.png
    const resp = await uploadAvatar(file);
    const url = (resp as any)?.url || (resp as any)?.data?.url || '';
    const path = (resp as any)?.path || (resp as any)?.data?.path || '';
    if (!url && !path) {
      emailErrorMsg.value = '上传失败，请稍后重试';
      showEmailError.value = true;
      return;
    }
    // 优先使用后端返回的完整URL；若缺失则拼接相对路径
    const fullUrl = url || `${API_URL}${path}`;
    // 后端 UploadAvatar 已写入数据库，无需再次调用 updateProfile，避免重复请求导致误报失败
    // 更新本地展示
    formData.avatar = fullUrl;
    // 同步更新全局 Store，确保其他页面（如对战页）能即时获取最新头像
    if (userStore.userInfo) {
      userStore.setUser({ ...userStore.userInfo, avatar: fullUrl });
    }
    showSaveSuccess.value = true;
    isDefaultAvatar.value = false;
  } catch (err: any) {
    // 更健壮的错误处理：拦截器可能返回字符串消息
    const msg = typeof err === 'string' ? err : (err?.response?.data?.message || err?.message || '上传失败，请重试');
    emailErrorMsg.value = msg;
    showEmailError.value = true;
  } finally {
    isUploadingAvatar.value = false;
    // 重置文件选择，防止同文件无法再次触发 change
    try { (e.target as HTMLInputElement).value = ''; } catch {}
  }
};

/**
 * @description: 退出登录
 */
const logout = async () => {
  showLogoutConfirm.value = false;

  try {
    // 先通知后端：将当前账号置为离线（online=0）
    try {
      await RequestHandler.post('/user/logout');
    } catch (e) {
      // 若因拦截器/token问题失败，不阻塞前端登出流程
    }
    // 尝试调用 store.logout，如果类型定义缺失则安全回退（as any）
    await (userStore as any).logout?.();

    // 额外确保本地存储的 token/userInfo 被清理（以防 store 未实现）
    try { localStorage.removeItem('token'); } catch {}
    try { localStorage.removeItem('userInfo'); } catch {}
    // 同步清理本地保存的登录凭据（Local Storage）
    // 这些字段通常在登录页用于“记住我”功能
    try { localStorage.removeItem('email'); } catch {}
    try { localStorage.removeItem('password'); } catch {}
    // 兼容性：若有存入 Session Storage，也一并清理
    try { sessionStorage.removeItem('email'); } catch {}
    try { sessionStorage.removeItem('password'); } catch {}

    // 使用 replace 防止用户按后退回到已登录页面
    await router.replace('/auth/login');
  } catch (e) {
    console.error('退出登录失败:', e);
    // 即便出错也跳转到登录页，确保用户被登出视图层面
    await router.replace('/auth/login');
  }
};

const promptLogout = () => { showLogoutConfirm.value = true; };
const cancelLogout = () => { showLogoutConfirm.value = false; };


/**
 * @description: 注销账号
 */
const deleteAccount = () => {
  showDeleteConfirm.value = true;
};

const confirmDeleteAccount = async () => {
  isProcessing.value = true;
  try {
    // 调用后端注销接口
    await RequestHandler.post('/user/delete_account');
    // 清除本地登录信息
    userStore.logout?.();
    localStorage.removeItem('token');
    localStorage.removeItem('userInfo');
    // 跳转到登录页
    await router.replace('/auth/login');
  } catch (error) {
    emailErrorMsg.value = '注销失败，请稍后重试';
    showEmailError.value = true;
  } finally {
    isProcessing.value = false;
    showDeleteConfirm.value = false;
  }
};

/**
 * @description: 组件挂载时初始化表单
 */
onMounted(() => {
  initForm();
});
</script>


