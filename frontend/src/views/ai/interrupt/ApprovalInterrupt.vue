<template>
    <el-card class="approval-interrupt">
        <template #header>
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                    <el-tag type="warning" size="small">{{ t('ai.interrupt.approval.title') }}</el-tag>
                    <span class="font-medium">{{ interruptData?.title }}</span>
                </div>
                <enum-tag v-if="isProcessed" :enums="InterruptAction" :value="currentAction" />
                <el-tag v-else type="warning" size="small">
                    {{ t('ai.interrupt.approval.pendingApproval') }}
                </el-tag>
            </div>
        </template>

        <div class="space-y-3">
            <!-- 描述信息 -->
            <div class="text-sm text-gray-600 dark:text-gray-400">
                {{ interruptData?.description }}
            </div>

            <!-- 工具信息 -->
            <div v-if="interruptData?.toolInfo" class="bg-gray-50 dark:bg-gray-800 rounded p-3">
                <div class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">{{ t('ai.interrupt.approval.toolInfo') }}</div>
                <div class="space-y-2 text-sm">
                    <div class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.approval.name') }}:</span>
                        <span class="font-mono text-blue-600 dark:text-blue-400">{{ interruptData.toolInfo.name }}</span>
                    </div>
                    <div v-if="interruptData.toolInfo.desc" class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.approval.description') }}:</span>
                        <span class="text-gray-700 dark:text-gray-300 flex-1">{{ interruptData.toolInfo.desc }}</span>
                    </div>
                </div>
            </div>

            <!-- 参数信息 -->
            <div v-if="interruptData?.arguments" class="bg-gray-50 dark:bg-gray-800 rounded p-3">
                <div class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">{{ t('ai.interrupt.approval.executionParams') }}</div>
                <pre class="text-xs overflow-x-auto bg-white dark:bg-gray-900 p-2 rounded border border-gray-200 dark:border-gray-700">
                    {{ formatJson(interruptData.arguments) }}
                </pre>
            </div>

            <!-- 中断ID -->
            <div v-if="interruptId" class="text-xs text-gray-400 dark:text-gray-500">
                <span class="font-medium">{{ t('ai.interrupt.approval.interruptId') }}:</span>
                <span class="font-mono">{{ interruptId }}</span>
            </div>

            <!-- 操作结果记录 -->
            <div v-if="resumeInfo" class="bg-blue-50 dark:bg-blue-900/20 rounded p-3 border border-blue-200 dark:border-blue-800">
                <div class="text-xs font-medium text-blue-600 dark:text-blue-400 mb-2">{{ t('ai.interrupt.approval.operationRecord') }}</div>
                <div class="space-y-2 text-sm">
                    <div class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.approval.operationType') }}:</span>
                        <enum-tag v-if="isProcessed" :enums="InterruptAction" :value="resumeInfo.action" />
                    </div>
                    <div v-if="resumeInfo.timestamp" class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.approval.operationTime') }}:</span>
                        <span class="text-gray-700 dark:text-gray-300">{{ formatDate(resumeInfo.timestamp) }}</span>
                    </div>
                    <div v-if="resumeInfo.payload" class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.approval.additionalData') }}:</span>
                        <pre class="text-xs overflow-x-auto bg-white dark:bg-gray-900 p-2 rounded border border-gray-200 dark:border-gray-700 flex-1">{{
                            formatJson(resumeInfo.payload)
                        }}</pre>
                    </div>
                </div>
            </div>
        </div>

        <template #footer v-if="!readonly && !isProcessed">
            <div class="flex justify-end gap-2">
                <el-button size="small" @click="handleApprove">{{ t('ai.interrupt.approval.approve') }}</el-button>
                <el-button size="small" type="danger" @click="handleReject">{{ t('ai.interrupt.approval.reject') }}</el-button>
            </div>
        </template>
    </el-card>
</template>

<script setup lang="ts">
/**
 * 审批类型中断组件
 * 用于需要用户确认的高危操作场景
 */

import EnumValue from '@/common/Enum';
import { formatDate, formatJson } from '@/common/utils/format';
import EnumTag from '@/components/enumtag/EnumTag.vue';
import { ElMessageBox } from 'element-plus';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { InternalMessage, InterruptActionEvent } from './types';

interface Props {
    data: InternalMessage;
    readonly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
    readonly: false,
});

const emit = defineEmits<{
    action: [action: InterruptActionEvent];
}>();

const { t } = useI18n();

// 从 data 中提取常用字段
const interruptId = computed(() => props.data.actionId);
const turnId = computed(() => props.data.turnId);
const interruptData = computed(() => props.data.extra?.content);
const resumeInfo = computed(() => props.data.extra?.resumeInfo);
const currentAction = computed(() => resumeInfo.value?.action);
const isProcessed = computed(() => !!currentAction.value);

const InterruptAction = {
    Approve: EnumValue.of('approve', 'ai.interrupt.action.approve').tagTypeSuccess(),
    Reject: EnumValue.of('reject', 'ai.interrupt.action.reject').tagTypeDanger(),
};

/**
 * 处理审批通过操作
 */
const handleApprove = () => {
    handleAction('approve');
};

/**
 * 处理审批拒绝操作，弹出输入框
 */
const handleReject = async () => {
    try {
        const { value: reason } = await ElMessageBox.prompt(t('ai.interrupt.approval.rejectReasonPlaceholder'), t('ai.interrupt.approval.rejectTitle'), {
            confirmButtonText: t('common.confirm'),
            cancelButtonText: t('common.cancel'),
            inputType: 'textarea',
            inputPlaceholder: t('ai.interrupt.approval.rejectReasonPlaceholder'),
            inputValidator: (value) => {
                if (!value || !value.trim()) {
                    return t('ai.interrupt.approval.rejectReasonRequired');
                }
                return true;
            },
        });

        // 用户输入了拒绝原因，提交操作
        handleAction('reject', { reason: reason?.trim() });
    } catch {
        // 用户取消了操作，不做任何处理
    }
};

/**
 * 处理用户操作
 * @param action 操作类型
 * @param payload 额外数据
 */
const handleAction = (action: string, payload?: any) => {
    emit('action', {
        turnId: turnId.value || '',
        interruptId: interruptId.value || '',
        action,
        payload,
    });
};
</script>

<style scoped>
.approval-interrupt {
    @apply transition-all duration-300;
}
</style>
