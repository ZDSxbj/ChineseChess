<template>
  <div class="profile-container">
    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading">
      <div class="spinner"></div>
      <p>正在加载数据...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="hasError" class="error">
      <p>加载失败，请稍后重试。</p>
      <button @click="initForm()">重试</button>
    </div>

    <!-- 表单内容区 -->
    <form v-else class="profile-form" @submit.prevent>
      <div class="form-group">
        <label class="form-label">头像</label>
        <div class="avatar-upload">
          <!-- 如果没有头像则不渲染 img，保持空白 -->
          <img v-if="formData.avatar" :src="formData.avatar" alt="用户头像" class="avatar-img" />
          <div v-else class="avatar-empty"></div>

          <input type="file" id="avatar-input" class="avatar-input" @change="handleAvatarChange" />
          <label for="avatar-input" class="upload-btn">上传头像</label>
        </div>
      </div>

      <!-- 姓名：显示 + 箭头编辑 -->
      <div class="form-group">
        <label class="form-label">姓名</label>
        <div class="field-row">
          <div class="field-value">{{ formData.name }}</div>
          <button type="button" class="edit-arrow" @click="startEditing('name')">›</button>
        </div>
        <div v-if="editingField === 'name'" class="edit-input">
          <input type="text" class="form-control" v-model="tempValue" placeholder="请输入姓名" />
          <div class="field-actions">
            <button type="button" class="btn-save" @click="saveEdit('name')">保存</button>
            <button type="button" class="btn-cancel" @click="cancelEdit()">取消</button>
          </div>
        </div>
      </div>

      <!-- 性别 -->
      <div class="form-group">
        <label class="form-label">性别</label>
        <div class="field-row">
          <div class="field-value">{{ formData.gender }}</div>
          <button type="button" class="edit-arrow" @click="startEditing('gender')">›</button>
        </div>
        <div v-if="editingField === 'gender'" class="edit-input">
          <select class="form-control" v-model="tempValue">
            <option value="">请选择性别</option>
            <option value="男">男</option>
            <option value="女">女</option>
            <option value="其他">其他</option>
          </select>
          <div class="field-actions">
            <button type="button" class="btn-save" @click="saveEdit('gender')">保存</button>
            <button type="button" class="btn-cancel" @click="cancelEdit()">取消</button>
          </div>
        </div>
      </div>

      <!-- 邮箱 -->
      <div class="form-group">
        <label class="form-label">邮箱</label>
        <div class="field-row">
          <div class="field-value">{{ formData.email }}</div>
          <button type="button" class="edit-arrow" @click="startEditing('email')">›</button>
        </div>
        <div v-if="editingField === 'email'" class="edit-input">
          <input type="email" class="form-control" v-model="tempValue" placeholder="请输入邮箱" />
          <div class="field-actions">
            <button type="button" class="btn-save" @click="saveEdit('email')">保存</button>
            <button type="button" class="btn-cancel" @click="cancelEdit()">取消</button>
          </div>
        </div>
      </div>

      <!-- 经验值（只读，不可编辑） -->
      <div class="form-group">
        <label class="form-label">经验值</label>
        <div class="field-row">
          <div class="field-value">{{ formData.exp }}</div>
        </div>
      </div>

      <!-- 总场次（只读，不可编辑） -->
      <div class="form-group">
        <label class="form-label">总场次</label>
        <div class="field-row">
          <div class="field-value">{{ formData.totalGames }}</div>
        </div>
      </div>

      <!-- 胜率（只读，不可编辑） -->
      <div class="form-group">
        <label class="form-label">胜率</label>
        <div class="field-row">
          <div class="field-value">{{ formData.winRate }}%</div>
        </div>
      </div>

      <div class="form-actions">
        <button type="button" class="btn btn-secondary" @click="logout">退出登录</button>
        <button type="button" class="btn btn-danger" @click="deleteAccount" :disabled="isProcessing">
          <span v-if="!isProcessing">注销账号</span>
          <span v-else>处理中...</span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/store/useStore';
import { getProfile } from '@/api/user/getProfile';
import { updateProfile } from '@/api/user/updateProfile';

// 路由实例
const router = useRouter();
// 用户 Store
const userStore = useUserStore();

// 状态管理
const isLoading = ref(true);
const hasError = ref(false);
const isProcessing = ref(false); // 处理状态（用于注销账号）

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
        avatar: userData.avatar || '', // 不使用默认头像，缺省为空
        name: userData.name || '未设置姓名',
        gender: userData.gender || '未选择',
        email: userData.email || '未设置邮箱',
        exp: userData.exp || 0,
        totalGames: userData.totalGames || 0,
        winRate: userData.winRate || 0
      });
    } else {
      // 默认占位文本，avatar 保持空字符串
      Object.assign(formData, {
        avatar: '',
        name: '未设置姓名',
        gender: '未选择',
        email: '未设置邮箱',
        exp: 0,
        totalGames: 0,
        winRate: 0
      });
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
  // 仅允许编辑非数值字段（name/gender/email）
  if (key === 'exp' || key === 'totalGames' || key === 'winRate') return;
  editingField.value = key;
  // 复制当前显示值，取消不影响原值
  tempValue.value = formData[key];
};

const cancelEdit = () => {
  editingField.value = null;
  tempValue.value = null;
};

const saveEdit = async (key: string) => {
  // 简单校验示例
  if ((key === 'winRate' || key === 'exp' || key === 'totalGames') && (tempValue.value === '' || tempValue.value == null || isNaN(Number(tempValue.value)))) {
    alert('请输入有效数值');
    return;
  }
  try {
    // 调用更新API
    await updateProfile({ [key]: tempValue.value });
    // 更新本地显示
    formData[key] = tempValue.value;
    editingField.value = null;
    tempValue.value = null;
    alert('保存成功');
  } catch (error) {
    console.error('更新失败:', error);
    alert('保存失败，请重试');
  }
};

/**
 * @description: 处理头像上传
 */
const handleAvatarChange = (e: Event) => {
  const input = e.target as HTMLInputElement;
  if (input.files && input.files[0]) {
    const file = input.files[0];
    formData.avatar = URL.createObjectURL(file);
    // TODO: 实际项目中需要将文件上传到服务器
  }
};

/**
 * @description: 退出登录
 */
const logout = () => {
  if (confirm('确定要退出登录吗？')) {
    userStore.logout();
    router.push('/auth/login');
  }
};

/**
 * @description: 注销账号
 */
const deleteAccount = async () => {
  if (!confirm('警告：注销账号将删除所有个人数据，且无法恢复！确定要注销吗？')) {
    return;
  }

  try {
    isProcessing.value = true;

    // TODO: 替换为真实的注销账号 API 调用
    setTimeout(() => {
      userStore.logout(); // 清除登录状态
      alert('账号已成功注销');
      router.push('/auth/login'); // 跳转到登录页
      isProcessing.value = false;
    }, 1500);

  } catch (error) {
    console.error('注销账号失败:', error);
    alert('注销失败，请稍后重试');
    isProcessing.value = false;
  }
};

/**
 * @description: 组件挂载时初始化表单
 */
onMounted(() => {
  initForm();
});
</script>

<style scoped>
.profile-container {
  max-width: 800px;
  margin: 20px auto;
  padding: 30px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
}

/* 加载和错误状态 */
.loading, .error {
  text-align: center;
  padding: 50px 0;
  color: #666;
}
.spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 20px;
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-left-color: #666;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.error button {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #666;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

/* 表单样式 */
.profile-form {
  width: 80%;
  margin: 0 auto;
}
.form-group {
  margin-bottom: 25px;
  position: relative; /* 用于胜率百分比提示的定位 */
}
.form-label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #555;
}
.form-control {
  width: 100%;
  padding: 12px 15px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  transition: border-color 0.3s ease;
}
.form-control:focus {
  outline: none;
  border-color: #666;
  box-shadow: 0 0 0 2px rgba(102, 102, 102, 0.2);
}

/* 编辑行样式 */
.field-row { display:flex; align-items:center; gap:12px; }
.field-value { flex:1; color:#333; padding:8px 0; }
.edit-arrow {
  background:none;
  border:none;
  font-size:20px;
  cursor:pointer;
  color:#888;
  padding:6px 10px;
  border-radius:4px;
}
.edit-input { margin-top:10px; display:flex; align-items:center; gap:10px; }
.field-actions { display:flex; gap:8px; margin-left:auto; }
.btn-save { background:#28a745;color:#fff;border:none;padding:6px 10px;border-radius:4px;cursor:pointer; }
.btn-cancel { background:#ccc;color:#333;border:none;padding:6px 10px;border-radius:4px;cursor:pointer; }

/* 头像相关 */
.avatar-upload {
  display: flex;
  align-items: center;
  gap: 20px;
}
.avatar-img {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #eee;
}
/* 未上传头像时显示的空白占位（可自定义样式或保持透明） */
.avatar-empty {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: transparent;
  border: 2px dashed #f0f0f0;
}
.avatar-input {
  display: none; /* 隐藏原生文件输入框 */
}
.upload-btn {
  padding: 8px 16px;
  background-color: #f0f0f0;
  color: #333;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}
.upload-btn:hover {
  background-color: #e0e0e0;
}

/* 表单操作按钮 */
.form-actions {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-top: 40px;
}
.btn {
  padding: 12px 30px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  border: none;
  font-size: 16px;
}
.btn-secondary {
  background-color: #f0f0f0;
  color: #333;
}
.btn-secondary:hover {
  background-color: #e0e0e0;
}
.btn-danger {
  background-color: #ff4d4f;
  color: #fff;
}
.btn-danger:hover {
  background-color: #d9363e;
}
.btn-danger:disabled {
  background-color: #ffb4b4;
  cursor: not-allowed;
}
</style>