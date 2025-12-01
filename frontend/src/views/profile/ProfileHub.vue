<template>
  <div class="profile-hub-container">
    <div class="sidebar">
      <h2>个人中心</h2>
      <ul>
        <li>
          <RouterLink to="/profile/info" class="sidebar-link" active-class="active">
            个人信息
          </RouterLink>
        </li>
        <li>
          <RouterLink to="records" class="sidebar-link" active-class="active">
            行棋记录
          </RouterLink>
        </li>
      </ul>
    </div>
    <div class="content-area">
      <!-- 这里是子路由的出口，点击侧边栏链接后，对应的组件会在这里渲染 -->
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router';
</script>

<style scoped>
.profile-hub-container {
  display: flex;
  min-height: calc(100vh - 120px);
  background-color: #f5f5f5;
}

.sidebar {
  width: 220px;
  background-color: #fff;
  box-shadow: 2px 0 10px rgba(0, 0, 0, 0.05);
  padding: 30px 0;
}

.sidebar h2 {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
}

.sidebar ul {
  list-style: none;
  padding: 0;
}

.sidebar-link {
  display: block;
  padding: 15px 30px;
  color: #555;
  text-decoration: none;
  transition: all 0.3s ease;
}

.sidebar-link:hover {
  background-color: #f0f0f0;
  color: #000;
}

.sidebar-link.active {
  background-color: #42b983;
  color: #fff;
  font-weight: 500;
}

.content-area {
  flex: 1;
  padding: 30px;
  overflow-y: auto;
}
</style>