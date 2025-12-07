import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import UserGroups from '../views/UserGroups.vue'
import Users from '../views/Users.vue'
import OnlineUsers from '../views/OnlineUsers.vue'
import Logs from '../views/Logs.vue'

const routes = [
  { path: '/', component: Dashboard },
  { path: '/groups', component: UserGroups },
  { path: '/users', component: Users },
  { path: '/online', component: OnlineUsers },
  { path: '/logs', component: Logs }
]

export default createRouter({
  history: createWebHistory(),
  routes
})
