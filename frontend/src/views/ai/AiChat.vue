<template>
    <div class="h-full flex flex-col p-5 justify-center">
        <div class="w-full flex-1 overflow-hidden" v-if="state.messages.length > 0">
            <BubbleList v-loading="msgLoading" ref="bubbleListRef" :list="messages" max-height="100%">
                <template #avatar="{ item }">
                    <SvgIcon v-if="item.role == ROLE.AI" :size="24" name="icon ai/assistant" color="var(--el-color-primary)" />
                    <img v-else class="size-10 max-w-none rounded-full" :src="useUserInfo().userInfo.photo" alt="avatar" />
                </template>

                <template #header="{ item }">
                    <ThoughtChain :thinking-items="item.thinks" dot-size="small" class="min-w-150 max-w-300" :max-width="BUBBLE_MAX_WIDTH" row-key="id">
                    </ThoughtChain>

                    <!-- 中断类型：展示动态中断组件 -->
                    <div v-for="internal in item.internals" :key="internal.id || internal.extra?.interruptId" class="mb-2">
                        <!-- 只有 extra.type 为 interrupt 时才展示中断组件 -->
                        <component
                            class="w-200"
                            v-if="internal.extra?.type === 'interrupt'"
                            :is="getInterruptComponent(internal.extra.content?.type)"
                            :data="internal"
                            :readonly="internal.extra?.toolStatus !== 'interrupted'"
                            @action="handleInterruptAction"
                        />

                        <!-- 其他类型的 internal 可以在这里扩展 -->
                        <div
                            v-else-if="internal.extra?.type === 'notification'"
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
                        <div v-else class="p-3 bg-gray-50 dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700">
                            <div class="text-xs text-gray-500 dark:text-gray-400">
                                <span class="font-medium">类型:</span> {{ internal.extra?.type || internal.type || 'unknown' }}
                            </div>
                            <div v-if="internal.content || internal.extra?.content" class="text-xs text-gray-600 dark:text-gray-300 mt-1">
                                {{ internal.content || internal.extra?.content?.description || JSON.stringify(internal.extra?.content || internal.content) }}
                            </div>
                            <!-- 显示原始数据用于调试 -->
                            <details class="mt-2 text-xs">
                                <summary class="cursor-pointer text-gray-400">查看完整数据</summary>
                                <pre class="mt-1 p-2 bg-white dark:bg-gray-900 rounded overflow-x-auto">{{ JSON.stringify(internal, null, 2) }}</pre>
                            </details>
                        </div>
                    </div>
                </template>

                <template #content="{ item }">
                    <!-- chat 内容走 markdown -->
                    <XMarkdown
                        v-if="item.role === ROLE.AI || item.role == ROLE.INTERNAL"
                        :markdown="item.content"
                        :themes="{ light: 'github-light', dark: 'github-dark' }"
                        :default-theme-mode="isDark ? 'dark' : 'light'"
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
            <EditorSender
                ref="senderRef"
                @click.once="onFoucsSender()"
                style="border-radius: 24px"
                :custom-style="{
                    height: '60px',
                }"
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
            </EditorSender>

            <el-dialog v-model="dialogCustomVisible" title="自定义触发符号选择弹窗" width="500">
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
import { computed, onBeforeUnmount, reactive, ref, toRefs, useTemplateRef, watch } from 'vue';
import { BubbleList, EditorSender, ThoughtChain, XMarkdown } from 'vue-element-plus-x';
import type { BubbleListInstance } from 'vue-element-plus-x/types/BubbleList';
import type { SubmitResult } from 'vue-element-plus-x/types/EditorSender';
import { useI18n } from 'vue-i18n';
import { aiApi, SessionMessage, ToolCall } from './api';
import { getInterruptComponent } from './interrupt';
import { InterruptActionEvent } from './interrupt/types';
import { ElMessage } from 'element-plus';

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
};

const props = defineProps({
    sessionId: {
        type: String,
        default: '',
    },
    isNewSession: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(['activate']);

let socket: WebSocket;
let reconnectTimer: any = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000;

const themeConfig = useThemeConfig();
const isDark = computed(() => themeConfig.themeConfig.isDark);

const senderRef = useTemplateRef<InstanceType<typeof EditorSender>>('senderRef');
const bubbleListRef = useTemplateRef<BubbleListInstance>('bubbleListRef');

const state = reactive({
    msgLoading: false,
    senderLoading: false,
    dialogCustomVisible: false,
    tagPrefix: '',
    messages: [] as Array<messageType>,
});
const { msgLoading, senderLoading, dialogCustomVisible } = toRefs(state);

// 标记会话是否已激活（用户是否发送过消息）
const sessionActivated = ref(false);

const initSocket = async () => {
    try {
        console.log('init chat ws...');
        socket = await createWebSocket(`/ai/chat`);
        socket.onmessage = (e) => {
            const data: SessionMessage = JSON.parse(e.data);
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
        dialogTitle: '资产选择',
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
 * @param action 中断操作事件数据
 */
const handleInterruptAction = async (action: InterruptActionEvent) => {
    console.log('中断操作:', action);
    sendUserMsg('interruptResume', JSON.stringify(action));
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
const appendToolResult = (message: messageType, actionId: string, content: string) => {
    if (!message.thinks) {
        return;
    }

    for (let think of message.thinks) {
        if (think.type !== THINK_TYPE.TOOL || !think.extra) {
            continue;
        }
        if (think.extra.toolCallId === actionId) {
            think.thinkContent = content ? `${think.thinkContent}  ${t('ai.chat.toolCallResult')}: ${content}` : content;
            console.log(content);
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
        return;
    }

    if (!message.internals) {
        message.internals = [];
    }
    message.internals.push(internalMsg as InternalMessageType);
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
    const targetInterrupt = message.internals.find((internal) => internal.actionId === interruptId && internal.extra?.type === 'interrupt');

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
    if (props.isNewSession && !sessionActivated.value) {
        sessionActivated.value = true;
        setTimeout(() => {
            emit('activate');
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
    scrollToBottom();

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
            if (message.actionId) {
                appendToolResult(targetMessage, message.actionId, message.content || '');
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

    // 使用统一的消息处理器
    processMessage(chunkMsg, message);
};

const loadMessage = async () => {
    if (!props.sessionId) {
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
            toolResults: Map<string, SessionMessage>;
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
        if (message.role === ROLE.TOOL && message.actionId) {
            group.toolResults.set(message.actionId, message);
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
            continue;
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

    console.log('Converted messages:', finalMessages);
    return finalMessages;
};

const onFoucsSender = () => {
    senderRef.value?.focusToEnd();
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

onBeforeUnmount(() => {
    cleanupSocket();
});

watch(
    () => props.sessionId,
    async (newVal) => {
        if (!newVal) {
            return;
        }
        if (!socket) {
            initSocket();
        }
        if (props.isNewSession) {
            state.senderLoading = false;
        }
        onFoucsSender();
        loadMessage();
    },
    { immediate: true }
);

/**
 * 转为组件需要的数据
 */
const messages = computed(() => {
    return state.messages.map((item) => {
        const role = item.role;
        return {
            ...item,
            time: formatDate(item.time),
            placement: role === ROLE.AI || role == ROLE.INTERNAL ? 'start' : 'end',
            variant: role === ROLE.AI ? 'filled' : 'outlined', // 气泡的样式
            isFog: role === ROLE.AI, // AI 消息开启雾化效果
            maxWidth: BUBBLE_MAX_WIDTH,
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
                content: '正在重新连接服务器，请稍候...',
                role: ROLE.AI,
            });
            attemptReconnect();
            return;
        }

        // 立即尝试重连
        attemptReconnect();
        state.messages.push({
            content: '连接已断开，正在尝试重新连接...',
            role: ROLE.AI,
        });
        return;
    }

    socket.send(
        JSON.stringify({
            type,
            sessionId: props.sessionId,
            content,
        })
    );
};

const onSubmit = (value: SubmitResult) => {
    try {
        state.senderLoading = true;
        sendUserMsg('text', value.text);

        state.messages.push({
            content: value.text,
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
