<template>
    <el-card class="generic-interrupt">
        <template #header>
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                    <el-tag type="info" size="small">{{ interruptData?.type || t('ai.interrupt.generic.interrupt') }}</el-tag>
                    <span class="font-medium">{{ interruptData?.title || t('ai.interrupt.generic.operationInterrupted') }}</span>
                </div>
                <el-tag v-if="isProcessed" :type="getActionTag(currentAction)" size="small">
                    {{ getActionText(currentAction) }}
                </el-tag>
                <el-tag v-else type="warning" size="small">
                    {{ t('ai.interrupt.generic.pending') }}
                </el-tag>
            </div>
        </template>

        <div class="space-y-3">
            <!-- 描述信息 -->
            <div v-if="interruptData?.description" class="text-sm text-gray-600 dark:text-gray-400">
                {{ interruptData.description }}
            </div>

            <!-- 原始数据展示（调试用） -->
            <div class="bg-gray-50 dark:bg-gray-800 rounded p-3">
                <div class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">{{ t('ai.interrupt.generic.details') }}</div>
                <pre class="text-xs overflow-x-auto bg-white dark:bg-gray-900 p-2 rounded border border-gray-200 dark:border-gray-700">{{
                    formatJson(interruptData)
                }}</pre>
            </div>

            <!-- 中断ID -->
            <div v-if="interruptId" class="text-xs text-gray-400 dark:text-gray-500">
                <span class="font-medium">{{ t('ai.interrupt.generic.interruptId') }}:</span>
                <span class="font-mono">{{ interruptId }}</span>
            </div>

            <!-- 操作结果记录 -->
            <div v-if="resumeInfo" class="bg-blue-50 dark:bg-blue-900/20 rounded p-3 border border-blue-200 dark:border-blue-800">
                <div class="text-xs font-medium text-blue-600 dark:text-blue-400 mb-2">{{ t('ai.interrupt.generic.operationRecord') }}</div>
                <div class="space-y-2 text-sm">
                    <div class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.generic.operationType') }}:</span>
                        <el-tag :type="getActionTag(resumeInfo.action)" size="small">
                            {{ getActionText(resumeInfo.action) }}
                        </el-tag>
                    </div>
                    <div v-if="resumeInfo.timestamp" class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.generic.operationTime') }}:</span>
                        <span class="text-gray-700 dark:text-gray-300">{{ formatDate(resumeInfo.timestamp) }}</span>
                    </div>
                    <div v-if="resumeInfo.payload" class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.generic.additionalData') }}:</span>
                        <pre class="text-xs overflow-x-auto bg-white dark:bg-gray-900 p-2 rounded border border-gray-200 dark:border-gray-700 flex-1">{{
                            formatJson(resumeInfo.payload)
                        }}</pre>
                    </div>
                </div>
            </div>
        </div>

        <template #footer v-if="!readonly && !isProcessed">
            <div class="flex justify-end gap-2">
                <el-button size="small" @click="handleAction('approve')">{{ t('ai.interrupt.generic.confirm') }}</el-button>
                <el-button size="small" type="danger" @click="handleAction('reject')">{{ t('ai.interrupt.generic.reject') }}</el-button>
            </div>
        </template>
    </el-card>
</template>

<script setup lang="ts">
/**
 * 通用中断组件
 * 用于未注册特定类型的中断场景，作为降级方案
 */

import { formatDate, formatJson } from '@/common/utils/format';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { InternalMessage, InterruptActionEvent } from './types';

const { t } = useI18n();

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

// 从 data 对象中提取常用字段
const interruptData = computed(() => props.data.extra?.content);
const interruptId = computed(() => props.data?.actionId);
const turnId = computed(() => props.data.turnId);
const resumeInfo = computed(() => props.data.extra?.resumeInfo);

// 根据 resumeInfo.action 计算当前动作
const currentAction = computed(() => resumeInfo.value?.action);

// 判断是否已处理（有 resumeInfo 表示已处理）
const isProcessed = computed(() => !!resumeInfo.value);

/**
 * 处理用户操作
 */
const handleAction = (action: string, payload?: any) => {
    emit('action', {
        turnId: turnId.value || '',
        interruptId: interruptId.value || '',
        action,
        payload,
    });
};

/**
 * 获取操作类型对应的标签类型
 */
const getActionTag = (actionType: string): 'success' | 'danger' | 'info' => {
    switch (actionType) {
        case 'approve':
        case 'confirm':
            return 'success';
        case 'reject':
        case 'cancel':
            return 'danger';
        default:
            return 'info';
    }
};

/**
 * 获取操作类型的显示文本
 */
const getActionText = (action: string): string => {
    switch (action) {
        case 'approve':
            return t('ai.interrupt.action.approve');
        case 'reject':
            return t('ai.interrupt.action.reject');
        case 'confirm':
            return t('ai.interrupt.action.confirm');
        case 'cancel':
            return t('ai.interrupt.action.cancel');
        default:
            return action;
    }
};
</script>

<style scoped>
.generic-interrupt {
    @apply transition-all duration-300;
}
</style>
