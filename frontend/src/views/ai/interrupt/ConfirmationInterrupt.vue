<template>
    <el-card class="confirmation-interrupt">
        <template #header>
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                    <el-tag type="primary" size="small">{{ t('ai.interrupt.confirmation.title') }}</el-tag>
                    <span class="font-medium">{{ interruptData?.title }}</span>
                </div>
                <enum-tag v-if="isProcessed" :enums="InterruptAction" :value="currentAction" />
                <el-tag v-else type="warning" size="small">
                    {{ t('ai.interrupt.confirmation.pendingConfirmation') }}
                </el-tag>
            </div>
        </template>

        <div class="space-y-3">
            <!-- 描述信息 -->
            <div class="text-sm text-gray-600 dark:text-gray-400">
                {{ interruptData?.description }}
            </div>

            <!-- 确认选项 -->
            <div v-if="interruptData?.options" class="bg-blue-50 dark:bg-blue-900/20 rounded p-3 border border-blue-200 dark:border-blue-800">
                <div class="text-xs font-medium text-blue-600 dark:text-blue-400 mb-2">{{ t('ai.interrupt.confirmation.pleaseSelect') }}</div>
                <el-radio-group v-model="selectedOption" :disabled="readonly || isProcessed">
                    <el-radio v-for="option in interruptData.options" :key="option.value" :value="option.value" class="block mb-2">
                        {{ option.label }}
                    </el-radio>
                </el-radio-group>
            </div>

            <!-- 中断ID -->
            <div v-if="interruptId" class="text-xs text-gray-400 dark:text-gray-500">
                <span class="font-medium">{{ t('ai.interrupt.confirmation.interruptId') }}:</span>
                <span class="font-mono">{{ interruptId }}</span>
            </div>

            <!-- 操作结果记录 -->
            <div v-if="resumeInfo" class="bg-blue-50 dark:bg-blue-900/20 rounded p-3 border border-blue-200 dark:border-blue-800">
                <div class="text-xs font-medium text-blue-600 dark:text-blue-400 mb-2">{{ t('ai.interrupt.confirmation.operationRecord') }}</div>
                <div class="space-y-2 text-sm">
                    <div class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.confirmation.operationType') }}:</span>
                        <enum-tag v-if="isProcessed" :enums="InterruptAction" :value="resumeInfo.action" />
                    </div>
                    <div v-if="resumeInfo.timestamp" class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.confirmation.operationTime') }}:</span>
                        <span class="text-gray-700 dark:text-gray-300">{{ formatDate(resumeInfo.timestamp) }}</span>
                    </div>
                    <div v-if="resumeInfo.payload" class="flex items-start gap-2">
                        <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ t('ai.interrupt.confirmation.selectionResult') }}:</span>
                        <span class="text-gray-700 dark:text-gray-300">{{ resumeInfo.payload }}</span>
                    </div>
                </div>
            </div>
        </div>

        <template #footer v-if="!readonly && !isProcessed">
            <div class="flex justify-end gap-2">
                <el-button size="small" type="primary" @click="handleAction('confirm', selectedOption)" :disabled="!selectedOption">
                    {{ t('ai.interrupt.confirmation.confirm') }}
                </el-button>
                <el-button size="small" @click="handleAction('cancel')">{{ t('ai.interrupt.confirmation.cancel') }}</el-button>
            </div>
        </template>
    </el-card>
</template>

<script setup lang="ts">
/**
 * 确认类型中断组件
 * 用于需要用户从多个选项中选择的场景
 */

import { EnumValue } from '@/common/Enum';
import { formatDate } from '@/common/utils/format';
import EnumTag from '@/components/enumtag/EnumTag.vue';
import { computed, ref } from 'vue';
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

const selectedOption = ref<string>();

// 从 data 对象中提取常用字段
const interruptData = computed(() => props.data.extra?.content);
const interruptId = computed(() => props.data.actionId);
const turnId = computed(() => props.data.turnId);
const resumeInfo = computed(() => props.data.extra?.resumeInfo);

// 根据 resumeInfo.action 计算当前状态
const currentAction = computed(() => resumeInfo.value?.action);

// 判断是否已处理（有 resumeInfo 表示已处理）
const isProcessed = computed(() => !!resumeInfo.value);

const InterruptAction = {
    Confirm: EnumValue.of('confirm', 'ai.interrupt.action.confirm').tagTypeSuccess(),
    Cancel: EnumValue.of('cancel', 'ai.interrupt.action.cancel').tagTypeDanger(),
};

/**
 * 处理用户操作
 * @param action 操作类型
 * @param payload 额外数据（如选中的选项值）
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
.confirmation-interrupt {
    @apply transition-all duration-300;
}
</style>
