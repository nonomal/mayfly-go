<template>
    <div class="h-full flex flex-col p-5 justify-center">
        <div class="w-full flex-1 overflow-hidden" v-if="state.messages.length > 0">
            <BubbleList v-loading="msgLoading" ref="bubbleListRef" :key="bubbleListKey" :list="messages" max-height="100%" :virtual="false">
                <template #avatar="{ item }">
                    <SvgIcon v-if="item.role == ROLE.AI" :size="24" name="icon ai/assistant" color="var(--el-color-primary)" />
                    <img v-else class="size-10 max-w-none rounded-full" :src="useUserInfo().userInfo.photo" alt="avatar" />
                </template>

                <template #header="{ item }">
                    <ThoughtChain :thinking-items="item.thinks" dot-size="small" class="min-w-150 max-w-300" :max-width="BUBBLE_MAX_WIDTH" row-key="id">
                    </ThoughtChain>

                    <!-- 中断组件：横向 flex 排列，宽度缩小 -->
                    <div class="flex flex-wrap gap-2">
                        <template v-for="internal in item.internals" :key="internal.id || internal.extra?.interruptId">
                            <component
                                class="max-w-120 flex-shrink-0"
                                v-if="internal.extra?.type?.startsWith('interrupt_')"
                                :is="getInterruptComponent(internal.extra?.type)"
                                :data="internal"
                                :readonly="internal.extra?.resumeInfo"
                                @action="handleInterruptAction"
                            />
                        </template>
                    </div>

                    <!-- 其他类型 internal 纵向排列 -->
                    <div v-for="internal in item.internals" :key="internal.id || internal.extra?.interruptId" class="mt-1">
                        <div
                            v-if="internal.extra?.type === 'notification'"
                            class="p-3 bg-blue-50 dark:bg-blue-900/20 rounded border border-blue-200 dark:border-blue-800"
                        >
                            <div class="text-sm font-medium text-blue-700 dark:text-blue-300">
                                {{ internal.content?.title || internal.extra?.content?.title }}
                            </div>
                            <div class="text-xs text-blue-600 dark:text-blue-400 mt-1">
                                {{ internal.content?.description || internal.extra?.content?.description }}
                            </div>
                        </div>

                        <!-- 未知类型的降级展示 -->
                        <div
                            v-else-if="!internal.extra?.type?.startsWith('interrupt_')"
                            class="px-2 py-1.5 bg-gray-50 dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700"
                        >
                            <div class="text-xs text-gray-500 dark:text-gray-400">
                                <span class="font-medium">{{ t('ai.chat.type') }}:</span> {{ internal.extra?.type || internal.type || 'unknown' }}
                            </div>
                            <div v-if="internal.content || internal.extra?.content" class="text-xs text-gray-600 dark:text-gray-300 mt-0.5">
                                {{ internal.content || internal.extra?.content?.description || JSON.stringify(internal.extra?.content || internal.content) }}
                            </div>
                        </div>
                    </div>

                    <!-- 中断处理进度提示：当有未处理的中断时显示进度 -->
                    <div
                        v-if="item.unprocessedInterruptCount > 0"
                        class="mt-2 px-2 py-1 bg-blue-50 dark:bg-blue-900/20 rounded border border-blue-200 dark:border-blue-800"
                    >
                        <div class="flex items-center justify-between">
                            <span class="text-xs text-blue-700 dark:text-blue-300">
                                {{ t('ai.chat.processed') }} {{ item.pendingResumes?.length || 0 }} / {{ item.unprocessedInterruptCount }}
                                {{ t('ai.chat.interrupts') }}
                                <span
                                    v-if="item.pendingResumes?.length === item.unprocessedInterruptCount"
                                    class="ml-1 text-xs text-green-600 dark:text-green-400"
                                >
                                    {{ t('ai.chat.submitting') }}
                                </span>
                            </span>
                        </div>
                    </div>
                </template>

                <template #content="{ item }">
                    <!-- chat 内容走 markdown -->
                    <MarkdownRenderer
                        v-if="item.role === ROLE.AI || item.role == ROLE.INTERNAL"
                        :markdown="item.content"
                        :enable-animate="true"
                        :is-dark="isDark"
                        :themes="{ light: 'github-light', dark: 'github-dark' }"
                        :default-theme-mode="isDark ? 'dark' : 'light'"
                        class="max-w-300"
                    />

                    <!-- user 内容 纯文本 -->
                    <div v-if="item.role === ROLE.USER" class="whitespace-pre-wrap">
                        {{ item.content }}
                    </div>
                </template>

                <template #footer="{ item }">
                    <div class="flex justify-between items-center">
                        <div>
                            <el-button @click="copyToClipboard(item.content)" color="#626aef" icon="DocumentCopy" size="small" circle />
                        </div>
                        <div class="ml-2 mt-1 text-xs">
                            {{ item.time }}
                        </div>
                    </div>
                </template>
            </BubbleList>
        </div>

        <!-- 输入框区域：固定在底部 -->
        <div class="w-full mt-4">
            <XSender
                ref="senderRef"
                @click.once="onFoucsSender()"
                style="border-radius: 24px"
                @submit="onSubmit"
                :loading="senderLoading"
                :disabled="senderLoading"
                :custom-dialog="true"
                :custom-trigger="customSenderTrigger"
                @show-tag-dialog="showTagDialog"
                submit-type="enter"
                :auto-focus="true"
                variant="updown"
                clearable
            >
            </XSender>

            <el-dialog v-model="dialogCustomVisible" :title="t('ai.chat.customTriggerDialogTitle')" width="500">
                <template v-for="option of customSenderTrigger" :key="option.prefix">
                    <p v-for="tag of option.tagList" :key="tag.id" @click="checkTag(tag)">
                        {{ tag.name }}
                    </p>
                </template>
            </el-dialog>
        </div>
    </div>
</template>

<script setup lang="ts" name="AiChat">
import { createWebSocket } from '@/common/request';
import { formatDate } from '@/common/utils/format';
import { copyToClipboard } from '@/common/utils/string';
import { useThemeConfig } from '@/store/themeConfig';
import { useUserInfo } from '@/store/userInfo';
import { computed, onBeforeUnmount, provide, reactive, ref, toRefs, useTemplateRef, watch } from 'vue';
import { BubbleList, ThoughtChain, XSender } from 'vue-element-plus-x';
import type { BubbleListInstance } from 'vue-element-plus-x/types/BubbleList';
import { useI18n } from 'vue-i18n';
import { aiApi, SessionMessage, ToolCall } from './api';
import { getInterruptComponent } from './interrupt';
import { InterruptActionEvent } from './interrupt/types';
import { ElMessage } from 'element-plus';
import { MarkdownRenderer } from 'x-markdown-vue';
import 'x-markdown-vue/style';

const { t } = useI18n();

// ==================== 常量定义 ====================

const ROLE = {
    AI: 'assistant',
    USER: 'user',
    TOOL: 'tool',
    INTERNAL: 'internal',
} as const;

const MESSAGE_TYPE = {
    END: 'end',
    ERROR: 'error',
} as const;

const THINK_STATUS = {
    LOADING: 'loading',
    SUCCESS: 'success',
    ERROR: 'error',
} as const;

const THINK_TYPE = {
    REASONING: 'reasoning',
    TOOL: 'tool',
} as const;

const BUBBLE_MAX_WIDTH = '80%';
const TOOL_ERROR_PREFIX = '[tool error]';

/**
 * Internal 消息类型
 */
type InternalMessageType = SessionMessage & {
    id?: string; // 内部消息ID
    actionId: string;
    extra?: {
        type?: 'interrupt' | 'resume' | string; // 内部消息类型
        [key: string]: any;
    };
};

type messageType = SessionMessage & {
    key?: number;
    loading?: boolean; // 是否正在加载中
    thinks?: Array<{
        type?: string;
        status: string;
        id: string;
        title: string;
        isCanExpand: boolean;
        isDefaultExpand: boolean;
        thinkTitle: string;
        thinkContent: string;
        extra?: any;
    }>; // 思考链

    internals?: Array<InternalMessageType>; // 内部消息数组
    pendingResumes?: InterruptActionEvent[]; // 待批量提交的中断恢复信息
};

const props = defineProps({
    sessionId: {
        type: String,
        default: '',
    },
});

const emit = defineEmits(['activate']);

// 内部跟踪当前 sessionId，支持新会话时从空值被后端赋值
const currentSessionId = ref('');

// 新会话标识（统一用 '-1' 表示）
const isNewSession = computed(() => props.sessionId === '-1');

// 标记会话是否已激活（用户是否发送过消息）
const sessionActivated = ref(false);

let socket: WebSocket;
let reconnectTimer: any = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000;

const themeConfig = useThemeConfig();
const isDark = computed(() => themeConfig.themeConfig.isDark);

const senderRef = useTemplateRef<InstanceType<typeof XSender>>('senderRef');
const bubbleListRef = useTemplateRef<BubbleListInstance>('bubbleListRef');

const state = reactive({
    msgLoading: false,
    senderLoading: false,
    dialogCustomVisible: false,
    tagPrefix: '',
    messages: [] as Array<messageType>,
    // 用于强制触发 messages computed 重新计算的版本号
    // 当修改消息内部状态（如 pendingResumes、internals）时递增
    version: 0,
});
const { msgLoading, senderLoading, dialogCustomVisible } = toRefs(state);

const initSocket = async () => {
    try {
        console.log('init chat ws...');
        socket = await createWebSocket(`/ai/chat`);
        socket.onmessage = (e) => {
            const data: SessionMessage = JSON.parse(e.data);

            // 会话隔离：只处理属于当前激活会话的消息
            if (data.sessionId && data.sessionId !== currentSessionId.value) {
                // 新会话首次收到后端返回的真实 sessionId，更新并通知父组件
                if (isNewSession.value) {
                    currentSessionId.value = data.sessionId;
                } else {
                    console.log(`忽略不属于当前会话的消息: ${data.sessionId} !== ${currentSessionId.value}`);
                    return;
                }
            }

            handleChunk(data);
        };

        socket.onclose = (event) => {
            console.log('chat ws 连接关闭:', event.code, event.reason);
            if (!event.wasClean) {
                attemptReconnect();
            }
        };

        socket.onerror = (error) => {
            console.error('chat ws  错误:', error);
        };

        // 连接成功，重置重连计数
        reconnectAttempts = 0;
    } catch (e) {
        state.messages.push({
            content: t('ai.chat.connectionFailed'),
            role: ROLE.AI,
        });
        console.log('连接错误', e);
        attemptReconnect();
        return;
    }
};

const attemptReconnect = () => {
    if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        console.warn('达到最大重连次数，停止重连');
        state.messages.push({
            content: t('ai.chat.connectionDisconnected'),
            role: ROLE.AI,
        });
        return;
    }

    reconnectAttempts++;
    console.log(`尝试第 ${reconnectAttempts} 次重连...`);

    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
    }

    reconnectTimer = setTimeout(() => {
        initSocket();
    }, RECONNECT_DELAY);
};

const cleanupSocket = () => {
    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
    }
    if (socket) {
        socket.onclose = null;
        socket.onerror = null;
        socket.onmessage = null;
        if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
            socket.close();
        }
        socket = null as any;
    }
    reconnectAttempts = 0;
};

// 自定义触发符相关逻辑
const customSenderTrigger = ref([
    {
        dialogTitle: t('ai.interrupt.assetSelection.title'),
        prefix: '#',
        tagList: [
            { id: 'ht1', name: '话题一' },
            { id: 'ht2', name: '话题二' },
        ],
    },
]);

const showTagDialog = (prefix: string) => {
    state.tagPrefix = prefix;
    state.dialogCustomVisible = true;
};

const checkTag = (tag: any) => {
    senderRef.value?.customSetTag(state.tagPrefix, tag);
    dialogCustomVisible.value = false;
};

/**
 * 统一的中断操作处理器
 * 所有中断操作均缓存到 pendingResumes，当所有中断都处理完毕后自动批量提交
 */
const handleInterruptAction = async (action: InterruptActionEvent) => {
    console.log('中断操作:', action);

    // 先尝试通过 turnId 找到 AI/INTERNAL 消息（中断只属于 AI 消息容器）
    let message = state.messages.find((m: messageType) => (m.role === ROLE.AI || m.role === ROLE.INTERNAL) && m.turnId && m.turnId === action.turnId);

    // 如果找不到，尝试通过 interruptId 找到包含该中断的 AI/INTERNAL 消息
    if (!message) {
        message = state.messages.find(
            (m: messageType) =>
                (m.role === ROLE.AI || m.role === ROLE.INTERNAL) &&
                m.internals?.some((internal: InternalMessageType) => (internal.actionId || internal.extra?.actionId) === action.interruptId)
        );
    }

    // 最后回退到最后一条 AI/INTERNAL 消息
    if (!message) {
        message = [...state.messages].reverse().find((m: messageType) => m.role === ROLE.AI || m.role === ROLE.INTERNAL);
    }

    if (!message) {
        console.warn('没有消息可处理中断');
        return;
    }

    // 初始化 pendingResumes
    if (!message.pendingResumes) {
        message.pendingResumes = [];
    }

    // 去重：同一 interruptId 只保留最新的
    const existingIndex = message.pendingResumes.findIndex((r: InterruptActionEvent) => r.interruptId === action.interruptId);
    if (existingIndex >= 0) {
        message.pendingResumes[existingIndex] = action;
    } else {
        message.pendingResumes.push(action);
    }

    // 更新对应 interrupt 的 pending 状态，用于 UI 显示
    const targetInterrupt = message.internals?.find(
        (internal: InternalMessageType) =>
            (internal.actionId || internal.extra?.actionId) === action.interruptId && internal.extra?.type?.startsWith('interrupt_')
    );
    if (targetInterrupt) {
        if (!targetInterrupt.extra) {
            targetInterrupt.extra = {};
        }
        targetInterrupt.extra.pendingResumeInfo = {
            action: action.action,
            payload: action.payload,
        };
    }

    // 递增版本号，强制触发 messages computed 重新计算
    state.version++;
    // 强制替换 messages 数组引用，确保 BubbleList 检测到变化
    state.messages = [...state.messages];

    // 检查是否所有中断都已处理，如果是则自动批量提交
    const unprocessedCount = (message.internals || []).filter(
        (internal: InternalMessageType) => internal.extra?.type?.startsWith('interrupt_') && !internal.extra?.resumeInfo
    ).length;

    if (message.pendingResumes && message.pendingResumes.length === unprocessedCount && unprocessedCount > 0) {
        // 所有中断已处理完毕，自动提交（setTimeout 让当前渲染周期完成）
        setTimeout(() => {
            submitPendingResumes(message.turnId || '');
        }, 0);
    }
};

// 提供中断操作处理器给子组件，绕过 Vue 动态组件事件传递问题
provide('handleInterruptAction', handleInterruptAction);

/**
 * 提交所有待处理的中断恢复信息（批量提交）
 * @param turnId 轮次ID
 */
const submitPendingResumes = (turnId: string) => {
    let message = state.messages.find((m: messageType) => (m.role === ROLE.AI || m.role === ROLE.INTERNAL) && m.turnId && m.turnId === turnId);
    if (!message) {
        message = [...state.messages].reverse().find((m: messageType) => m.role === ROLE.AI || m.role === ROLE.INTERNAL);
    }
    if (!message || !message.pendingResumes || message.pendingResumes.length === 0) {
        return;
    }

    // 发送批量恢复请求
    sendUserMsg('interruptResume', JSON.stringify(message.pendingResumes));

    // 将 pending 状态转为正式 resumeInfo，并清空 pending
    for (const resume of message.pendingResumes) {
        const targetInterrupt = message.internals?.find(
            (internal: InternalMessageType) =>
                (internal.actionId || internal.extra?.actionId) === resume.interruptId && internal.extra?.type?.startsWith('interrupt_')
        );
        if (targetInterrupt) {
            if (!targetInterrupt.extra) {
                targetInterrupt.extra = {};
            }
            targetInterrupt.extra.resumeInfo = {
                action: resume.action,
                payload: resume.payload,
            };
            delete targetInterrupt.extra.pendingResumeInfo;
        }
    }

    message.pendingResumes = [];

    // 递增版本号，强制触发 messages computed 重新计算
    state.version++;
    // 强制替换 messages 数组引用，确保 BubbleList 检测到变化
    state.messages = [...state.messages];
};

/**
 * 原子操作：追加工具调用到消息的思考链中
 * @param message 目标消息
 * @param toolCall 工具调用对象
 */
const appendToolCall = (message: messageType, toolCall: ToolCall) => {
    if (!message.thinks) {
        message.thinks = [];
    }

    // 检查是否已存在相同的 toolCall
    const existingThink = message.thinks.find((t) => t.type === THINK_TYPE.TOOL && t.extra?.toolCallId === toolCall.id);

    if (existingThink) {
        // 如果已存在，更新内容
        existingThink.thinkContent = JSON.stringify(toolCall);
    } else {
        // 否则创建新的 think
        message.thinks.push({
            type: THINK_TYPE.TOOL,
            status: THINK_STATUS.LOADING,
            id: String(toolCall.id),
            title: t('ai.chat.toolCall'),
            isCanExpand: true,
            isDefaultExpand: false,
            thinkTitle: t('ai.chat.toolCall') + ' - ' + toolCall.function?.name || '',
            thinkContent: JSON.stringify(toolCall),
            extra: {
                toolCallId: toolCall.id,
                toolName: toolCall.function?.name,
            },
        });
    }
};

/**
 * 原子操作：追加工具调用结果到消息的思考链中
 * @param message 目标消息
 * @param actionId 工具调用ID
 * @param content 工具执行结果
 */
const appendToolResult = (message: messageType, toolCallId: string, content?: string) => {
    if (!message.thinks || !content) {
        return;
    }

    for (let think of message.thinks) {
        if (think.type !== THINK_TYPE.TOOL || !think.extra) {
            continue;
        }
        if (think.extra.toolCallId === toolCallId) {
            think.thinkContent = content ? `${think.thinkContent}  ${t('ai.chat.toolCallResult')}: ${content}` : content;
            if (content.startsWith(TOOL_ERROR_PREFIX)) {
                think.status = THINK_STATUS.ERROR;
            } else {
                think.status = THINK_STATUS.SUCCESS;
            }
            return;
        }
    }
};

/**
 * 原子操作：追加内部消息到消息中
 * @param message 目标消息
 * @param internalMsg 内部消息
 */
const appendInternal = (message: messageType, internalMsg: SessionMessage) => {
    if (internalMsg.type == MESSAGE_TYPE.ERROR) {
        ElMessage.error(internalMsg.content);
        return;
    }
    // 处理会话结束或错误
    if (internalMsg.type == MESSAGE_TYPE.END) {
        handleTurnEnd(message, internalMsg);
        return;
    }

    // 处理 resume 类型的内部消息：合并到对应的 interrupt 中
    if (internalMsg.extra?.type === 'resume') {
        mergeResumeToInterrupt(message, internalMsg);
        scrollToBottom();
        return;
    }

    if (!message.internals) {
        message.internals = [];
    }
    message.internals.push(internalMsg as InternalMessageType);
    scrollToBottom();
};

/**
 * 将 resume 消息合并到对应的 interrupt 消息中
 * @param message 目标消息
 * @param resumeMsg resume 类型的内部消息
 */
const mergeResumeToInterrupt = (message: messageType, resumeMsg: SessionMessage) => {
    if (!message.internals || message.internals.length === 0) {
        console.warn('No internals found to merge resume message');
        return;
    }

    const resumeData = resumeMsg.extra?.content;
    const interruptId = resumeData?.interruptId;

    if (!interruptId) {
        console.warn('Resume message missing interruptId');
        return;
    }

    // 查找对应的 interrupt 消息
    const targetInterrupt = message.internals.find(
        (internal) => (internal.actionId || internal.extra?.actionId) === interruptId && internal.extra?.type?.startsWith('interrupt_')
    );

    if (!targetInterrupt) {
        console.warn(`Interrupt with id ${interruptId} not found`);
        return;
    }

    // 更新 interrupt 的状态和内容
    if (!targetInterrupt.extra) {
        targetInterrupt.extra = {};
    }

    // 记录操作信息
    targetInterrupt.extra.resumeInfo = {
        action: resumeData.action,
        timestamp: resumeMsg.time,
        payload: resumeData.payload,
    };

    console.log('Merged resume to interrupt:', {
        interruptId,
        action: resumeData.action,
    });
};

/**
 * 处理会话结束或错误
 */
const handleTurnEnd = (message: messageType, chunkMsg: SessionMessage) => {
    message.time = chunkMsg.time;

    if (message.content) {
        state.senderLoading = false;

        // 结束可能存在的思考链
        if (message.thinks && message.thinks.length > 0) {
            for (let think of message.thinks) {
                if (think.status == THINK_STATUS.LOADING && think.type == THINK_TYPE.REASONING) {
                    think.status = THINK_STATUS.SUCCESS;
                }
            }
        }
    }

    // 如果是新会话，触发会话激活事件
    if (isNewSession.value && !sessionActivated.value) {
        sessionActivated.value = true;
        setTimeout(() => {
            emit('activate', currentSessionId.value);
        }, 500);
    }
};

/**
 * 原子操作：追加推理内容到消息的思考链中
 * @param message 目标消息
 * @param reasoningContent 推理内容
 */
const appendReasoning = (message: messageType, reasoningContent: string) => {
    if (!reasoningContent) {
        return;
    }
    const title = t('ai.chat.thinking');

    const thinks = message.thinks;

    // 创建think对象的辅助函数
    const createThink = (status: string, id: number | string) => ({
        type: THINK_TYPE.REASONING,
        status,
        id: String(id),
        title: title,
        isCanExpand: true,
        isDefaultExpand: true,
        thinkTitle: title,
        thinkContent: reasoningContent,
    });

    // 如果没有thinks数组，初始化
    if (!thinks || thinks.length == 0) {
        message.thinks = [createThink(THINK_STATUS.LOADING, 1)];
        return;
    }

    const thinkIndex = thinks.length - 1;
    const think = thinks[thinkIndex];

    // 如果title不同，结束当前think并创建新的
    if (think.title != title) {
        thinks.push(createThink(THINK_STATUS.LOADING, thinkIndex + 2));
        return;
    }

    // 如果当前think还在loading状态，追加内容
    if (think.status == THINK_STATUS.LOADING) {
        think.thinkContent += reasoningContent;
        if (!reasoningContent) {
            think.status = THINK_STATUS.SUCCESS;
        }
        return;
    }

    // 否则创建新的think
    thinks.push(createThink(THINK_STATUS.LOADING, thinkIndex + 2));
};

/**
 * 原子操作：追加普通文本内容到消息
 * @param message 目标消息
 * @param content 文本内容
 */
const appendContent = (message: messageType, content: string) => {
    if (!content) {
        return;
    }
    message.content += content;

    if (message.loading) {
        message.loading = false;
    }
};

/**
 * 统一的消息处理器
 * 无论是 WebSocket chunk 还是历史消息，都通过此函数处理
 * @param message 要处理的消息对象
 * @param targetMessage 目标消息容器
 */
const processMessage = (message: SessionMessage, targetMessage: messageType) => {
    switch (message.role) {
        case ROLE.USER:
            // 用户消息：直接追加内容
            appendContent(targetMessage, message.content);
            break;

        case ROLE.AI:
            // AI 消息：处理内容、推理、工具调用
            appendContent(targetMessage, message.content);

            // 处理推理内容
            if (message.reasoningContent) {
                appendReasoning(targetMessage, message.reasoningContent);
            }

            // 处理工具调用
            if (message.toolCalls && message.toolCalls.length > 0) {
                for (let toolCall of message.toolCalls) {
                    appendToolCall(targetMessage, toolCall);
                }
            }
            break;

        case ROLE.TOOL:
            // 工具结果：更新对应的工具调用状态
            if (message.toolCallId) {
                appendToolResult(targetMessage, message.toolCallId, message.content || '');
            }
            break;

        case ROLE.INTERNAL:
            // 内部消息：添加到 internals 数组
            appendInternal(targetMessage, message);
            break;

        default:
            console.warn('Unknown message role:', message.role);
    }
};

const handleChunk = (chunkMsg: SessionMessage) => {
    const nowMsgIndex = state.messages.length - 1;
    const message = state.messages[nowMsgIndex];
    if (!message) {
        console.warn('No message to append chunk to: ', nowMsgIndex);
        return;
    }

    // 同步 turnId：后端返回的 chunk 消息携带 turnId，需要赋值给当前消息容器
    // 使用 != null 以支持空字符串 turnId（后端可能发送空字符串）
    if (chunkMsg.turnId != null && !message.turnId) {
        message.turnId = chunkMsg.turnId;
    }

    // 使用统一的消息处理器
    processMessage(chunkMsg, message);
};

const loadMessage = async () => {
    if (!props.sessionId || isNewSession.value) {
        return;
    }
    try {
        state.msgLoading = true;
        const messages = await aiApi.listMessages.request({ sessionKey: props.sessionId });
        state.messages = converterMessages(messages);
    } finally {
        state.msgLoading = false;
        scrollToBottom();
    }
};

const converterMessages = (messages: SessionMessage[]) => {
    // 按 turnId 分组消息
    const turnGroups = new Map<
        string,
        {
            userMessage?: messageType;
            aiMessage?: messageType;
            toolCalls: SessionMessage[];
            toolResults: Map<string, SessionMessage>; // toolCallId -> toolResult
            internals: SessionMessage[];
        }
    >();

    // 第一轮遍历：按 turnId 分类所有消息
    for (let message of messages) {
        if (!message.turnId) {
            console.warn('Message without turnId skipped:', message);
            continue;
        }

        if (!turnGroups.has(message.turnId)) {
            turnGroups.set(message.turnId, {
                userMessage: undefined,
                aiMessage: undefined,
                toolCalls: [],
                toolResults: new Map(),
                internals: [],
            });
        }

        const group = turnGroups.get(message.turnId)!;

        // 用户消息
        if (message.role === ROLE.USER) {
            group.userMessage = message;
            continue;
        }

        // 工具调用消息
        if (message.toolCalls && message.toolCalls.length > 0) {
            group.toolCalls.push(message);
            continue;
        }

        // 工具调用结果
        if (message.role === ROLE.TOOL && message.toolCallId) {
            group.toolResults.set(message.toolCallId, message);
            continue;
        }

        // AI 助手消息
        if (message.role === ROLE.AI) {
            group.aiMessage = message;
            continue;
        }

        // 内部消息（中断等）
        if (message.role === ROLE.INTERNAL) {
            group.internals.push(message);
            continue;
        }
    }

    // 第二轮遍历：构建最终消息列表，使用统一的消息处理逻辑
    const finalMessages: messageType[] = [];

    for (let [turnId, group] of turnGroups) {
        // 用户消息直接添加
        if (group.userMessage) {
            finalMessages.push({
                ...group.userMessage,
                time: formatDate(group.userMessage.time),
            });
        }

        // 创建主消息容器
        let mainMessage: messageType;

        if (group.aiMessage) {
            mainMessage = {
                ...group.aiMessage,
                loading: false,
                thinks: [],
                internals: [],
            };
        } else {
            // 如果没有 AI 消息，但有其他内容，创建一个 loading 状态的 AI 消息容器
            if (group.internals.length > 0 || group.toolCalls.length > 0 || group.toolResults.size > 0) {
                mainMessage = {
                    role: ROLE.AI,
                    content: '',
                    loading: true,
                    thinks: [],
                    internals: [],
                };
            } else {
                // 没有任何内容，跳过
                continue;
            }
        }

        // 处理工具调用消息
        for (let toolCallMsg of group.toolCalls) {
            processMessage(toolCallMsg, mainMessage);
        }

        // 处理工具调用结果
        for (let [, toolResultMsg] of group.toolResults) {
            processMessage(toolResultMsg, mainMessage);
        }

        // 处理内部消息
        for (let internalMsg of group.internals) {
            processMessage(internalMsg, mainMessage);
        }

        finalMessages.push(mainMessage);
    }

    return finalMessages;
};

const onFoucsSender = () => {
    // senderRef.value?.foucs();
};

/**
 * 滚动到底部
 * @param delay
 */
const scrollToBottom = (delay: number = 500) => {
    setTimeout(() => {
        bubbleListRef.value?.scrollToBottom();
    }, delay);
};

let isIniting = false;

const init = async () => {
    if (isIniting) {
        return;
    }
    isIniting = true;
    try {
        if (!socket) {
            initSocket();
        }
        if (isNewSession.value) {
            state.senderLoading = false;
        }
        onFoucsSender();
        loadMessage();
    } finally {
        isIniting = false;
    }
};

onBeforeUnmount(() => {
    cleanupSocket();
});

watch(
    () => props.sessionId,
    async (newVal: string) => {
        if (!newVal || newVal === currentSessionId.value) {
            return;
        }
        currentSessionId.value = newVal;
        if (isNewSession.value) {
            state.messages = [];
        }
        init();
    },
    { immediate: true }
);

/**
 * 转为组件需要的数据
 */
// BubbleList 强制重新渲染的 key，基于 version 变化
const bubbleListKey = computed(() => `bubble-list-${state.version}`);

const messages = computed(() => {
    return state.messages.map((item: messageType) => {
        const role = item.role;
        const unprocessedInterruptCount = (item.internals || []).filter(
            (internal: InternalMessageType) => internal.extra?.type?.startsWith('interrupt_') && !internal.extra?.resumeInfo
        ).length;
        const pendingCount = item.pendingResumes?.length || 0;
        // key 需包含 pendingCount 和 unprocessedInterruptCount，确保 BubbleList 在中断状态变化时重新渲染
        // param-completion 组件内部通过 watch 监听 pendingResumeInfo，可在重新挂载后恢复用户输入状态
        const key = item.key || `${item.turnId || 'no-turn'}-${pendingCount}-${unprocessedInterruptCount}`;
        return {
            key,
            role: item.role,
            content: item.content,
            time: formatDate(item.time),
            placement: role === ROLE.AI || role == ROLE.INTERNAL ? 'start' : 'end',
            variant: role === ROLE.AI ? 'filled' : 'outlined', // 气泡的样式
            isFog: role === ROLE.AI, // AI 消息开启雾化效果
            maxWidth: BUBBLE_MAX_WIDTH,
            thinks: item.thinks,
            internals: JSON.parse(JSON.stringify(item.internals || [])) as InternalMessageType[], // 深拷贝一份 internals，避免响应式问题
            loading: item.loading,
            extra: item.extra,
            turnId: item.turnId,
            reasoningContent: item.reasoningContent,
            pendingResumes: item.pendingResumes,
            unprocessedInterruptCount,
        };
    }) as any[];
});

const sendUserMsg = (type: 'text' | 'interruptResume', content: string) => {
    // 检查 WebSocket 连接状态
    if (!socket || socket.readyState === WebSocket.CLOSED || socket.readyState === WebSocket.CLOSING) {
        console.warn('WebSocket 连接已关闭，尝试重连...');

        // 如果正在重连中，等待重连完成
        if (reconnectAttempts > 0 && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            state.messages.push({
                content: t('ai.chat.reconnecting'),
                role: ROLE.AI,
            });
            attemptReconnect();
            return;
        }

        // 立即尝试重连
        attemptReconnect();
        state.messages.push({
            content: t('ai.chat.connectionLost'),
            role: ROLE.AI,
        });
        return;
    }

    socket.send(
        JSON.stringify({
            type,
            sessionId: currentSessionId.value,
            content,
        })
    );
};

const onSubmit = () => {
    try {
        state.senderLoading = true;
        const content = senderRef.value?.getModelValue().text;
        sendUserMsg('text', content);

        state.messages.push({
            content: content,
            role: ROLE.USER,
            time: new Date(),
        });

        state.messages.push({
            content: '',
            role: ROLE.AI,
            loading: true,
        });
    } finally {
        clearSenderEditor();
    }
};

const clearSenderEditor = () => {
    senderRef.value?.clear();
};
</script>

<style></style>
