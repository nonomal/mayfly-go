/**
 * 中断组件扩展示例
 * 
 * 本文件展示如何添加新的中断类型组件
 */

// ============================================
// 示例 1: 创建短信验证中断组件
// ============================================

/**
 * 文件: SmsVerificationInterrupt.vue
 * 
 * <template>
 *   <el-card class="sms-verification-interrupt">
 *     <template #header>
 *       <div class="flex items-center justify-between">
 *         <div class="flex items-center gap-2">
 *           <el-tag type="success" size="small">短信验证</el-tag>
 *           <span class="font-medium">{{ data.content?.title }}</span>
 *         </div>
 *         <el-tag :type="getStatusType(data.toolStatus)" size="small">
 *           {{ getStatusText(data.toolStatus) }}
 *         </el-tag>
 *       </div>
 *     </template>
 * 
 *     <div class="space-y-3">
 *       <div class="text-sm text-gray-600 dark:text-gray-400">
 *         {{ data.content?.description }}
 *       </div>
 * 
 *       <div v-if="data.content?.phone" class="bg-green-50 dark:bg-green-900/20 rounded p-3">
 *         <div class="text-xs font-medium text-green-600 dark:text-green-400 mb-2">验证码将发送至</div>
 *         <div class="text-lg font-mono text-green-700 dark:text-green-300">{{ data.content.phone }}</div>
 *       </div>
 * 
 *       <el-input
 *         v-model="verificationCode"
 *         placeholder="请输入6位验证码"
 *         maxlength="6"
 *         :disabled="readonly || data.toolStatus !== 'interrupted'"
 *       >
 *         <template #append>
 *           <el-button @click="sendSmsCode" :loading="sending">发送验证码</el-button>
 *         </template>
 *       </el-input>
 *     </div>
 * 
 *     <template #footer v-if="!readonly && data.toolStatus === 'interrupted'">
 *       <div class="flex justify-end gap-2">
 *         <el-button 
 *           type="primary" 
 *           size="small" 
 *           @click="$emit('verify', data.interruptId, verificationCode)"
 *           :disabled="!verificationCode || verificationCode.length !== 6"
 *         >
 *           验证
 *         </el-button>
 *         <el-button size="small" @click="$emit('cancel', data.interruptId)">取消</el-button>
 *       </div>
 *     </template>
 *   </el-card>
 * </template>
 * 
 * <script setup lang="ts">
 * import { ref } from 'vue';
 * 
 * interface Props {
 *   data: {
 *     interruptId: string;
 *     type: string;
 *     toolStatus: string;
 *     content: {
 *       title: string;
 *       description: string;
 *       phone?: string;
 *     };
 *   };
 *   readonly?: boolean;
 * }
 * 
 * const props = withDefaults(defineProps<Props>(), {
 *   readonly: false,
 * });
 * 
 * const emit = defineEmits<{
 *   verify: [interruptId: string, code: string];
 *   cancel: [interruptId: string];
 * }>();
 * 
 * const verificationCode = ref('');
 * const sending = ref(false);
 * 
 * const sendSmsCode = async () => {
 *   sending.value = true;
 *   // TODO: 调用发送验证码接口
 *   setTimeout(() => {
 *     sending.value = false;
 *   }, 1000);
 * };
 * 
 * const getStatusType = (status?: string) => {
 *   switch (status) {
 *     case 'verified': return 'success';
 *     case 'failed': return 'danger';
 *     case 'interrupted': return 'warning';
 *     default: return 'info';
 *   }
 * };
 * 
 * const getStatusText = (status?: string) => {
 *   switch (status) {
 *     case 'verified': return '已验证';
 *     case 'failed': return '验证失败';
 *     case 'interrupted': return '待验证';
 *     default: return status || '未知';
 *   }
 * };
 * </script>
 */

// ============================================
// 示例 2: 在 index.ts 中注册
// ============================================

/**
 * 在 src/views/ai/interrupt/index.ts 中添加：
 * 
 * import SmsVerificationInterrupt from './SmsVerificationInterrupt.vue';
 * 
 * const interruptComponentMap = new Map<string, Component>([
 *   ['APPROVAL', ApprovalInterrupt],
 *   ['CONFIRMATION', ConfirmationInterrupt],
 *   ['SMS_VERIFICATION', SmsVerificationInterrupt],  // 新增这一行
 * ]);
 */

// ============================================
// 示例 3: 后端返回数据格式
// ============================================

/**
 * {
 *   "extra": {
 *     "type": "SMS_VERIFICATION",
 *     "interruptId": "abc-123-def",
 *     "toolStatus": "interrupted",
 *     "content": {
 *       "title": "身份验证",
 *       "description": "为了保障账户安全，请完成短信验证",
 *       "phone": "138****8888"
 *     }
 *   }
 * }
 */

// ============================================
// 示例 4: 处理验证事件
// ============================================

/**
 * 在 AiChat.vue 或父组件中：
 * 
 * <component
 *   :is="getInterruptComponent(internal.extra?.type)"
 *   :data="internal.extra"
 *   @verify="handleSmsVerification"
 *   @cancel="handleCancel"
 * />
 * 
 * const handleSmsVerification = async (interruptId: string, code: string) => {
 *   try {
 *     // 调用后端验证接口
 *     await aiApi.verifySmsCode.request({
 *       interruptId,
 *       code
 *     });
 *     
 *     // 更新本地状态
 *     updateInterruptStatus(interruptId, 'verified');
 *   } catch (error) {
 *     ElMessage.error('验证失败');
 *   }
 * };
 */

export {};
