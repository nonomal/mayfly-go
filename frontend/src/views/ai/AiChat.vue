<template>
    <div class="h-full flex flex-col justify-center items-center p-5">
        <div class="w-full p-5" v-if="state.messages.length > 0" style="height: calc(100vh - 260px)">
            <BubbleList v-loading="msgLoading" ref="bubbleListRef" :list="messages" max-height="100%">
                <template #avatar="{ item }">
                    <SvgIcon v-if="item.role == ROLE.AI" :size="24" name="icon ai/assistant" color="var(--el-color-primary)" />
                    <img v-else class="size-10 max-w-none rounded-full" :src="useUserInfo().userInfo.photo" alt="avatar" />
                </template>

                <template #header="{ item }">
                    <ThoughtChain :thinking-items="item.thinks" class="min-w-150" :max-width="BUBBLE_MAX_WIDTH" row-key="codeId" />
                </template>

                <template #content="{ item }">
                    <!-- chat 内容走 markdown -->
                    <XMarkdown
                        v-if="item.content && item.role === ROLE.AI"
                        :markdown="item.content"
                        :themes="{ light: 'github-light', dark: 'github-dark' }"
                        :default-theme-mode="isDark ? 'dark' : 'light'"
                    />

                    <!-- user 内容 纯文本 -->
                    <div v-if="item.content && item.role === ROLE.USER" class="whitespace-pre-wrap">
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

        <div class="w-full">
            <EditorSender
                ref="senderRef"
                @click.once="onFoucsSender()"
                style="border-radius: 24px"
                :custom-style="{
                    height: '60px',
                }"
                @submit="onSubmit"
                :loading="senderLoading"
                submit-type="enter"
                :auto-focus="true"
                variant="updown"
                clearable
            >
            </EditorSender>
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
import { aiApi, SessionMessage, ToolCall } from './api';

// ==================== 常量定义 ====================

const ROLE = {
    AI: 'assistant',
    USER: 'user',
    TOOL: 'tool',
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

type messageType = SessionMessage & {
    key?: number;
    loading?: boolean; // 是否正在加载中
    thinks?: Array<{
        type?: string;
        status: string;
        codeId: string;
        title: string;
        isCanExpand: boolean;
        isDefaultExpand: boolean;
        thinkTitle: string;
        thinkContent: string;
        extra?: any;
    }>; // 思考链
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
    messages: [] as Array<messageType>,
});
const { msgLoading, senderLoading } = toRefs(state);

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
            content: '连接失败，请检查网络或联系管理员',
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
            content: '连接已断开，请刷新页面重试',
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

const handleChunk = (chunkMsg: SessionMessage) => {
    const nowMsgIndex = state.messages.length - 1;
    const message = state.messages[nowMsgIndex];

    if (chunkMsg.content && chunkMsg.role != ROLE.TOOL) {
        message.content += chunkMsg.content;
        scrollToBottom();

        if (message.loading) {
            message.loading = false;
        }
    }

    // 结束会话
    if (chunkMsg.type == MESSAGE_TYPE.END || chunkMsg.type == MESSAGE_TYPE.ERROR) {
        message.time = chunkMsg.time;
        state.senderLoading = false;

        // 结束可能存在的思考链
        if (message.thinks && message.thinks.length > 0) {
            for (let think of message.thinks) {
                if (think.status == THINK_STATUS.LOADING) {
                    think.status = THINK_STATUS.SUCCESS;
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
        return;
    }

    const reasoningChunk = chunkMsg.reasoningContent;
    if (reasoningChunk) {
        // 开始思考链状态
        addReasoningContent(nowMsgIndex, {
            title: '思考',
            reasoningContent: reasoningChunk,
        });
    }

    // 处理 <think> 标签格式的思考内容
    handleThinkTagContent(nowMsgIndex, chunkMsg.content);

    // 处理工具调用
    if (chunkMsg.toolCalls) {
        for (let toolCall of chunkMsg.toolCalls) {
            addReasoningContent(nowMsgIndex, {
                type: THINK_TYPE.TOOL,
                title: formatToolThinkTitle(toolCall),
                reasoningContent: JSON.stringify(toolCall),
                extra: {
                    toolCallId: toolCall.id,
                },
            });
        }
    }

    // 处理工具结果
    if (chunkMsg.role == ROLE.TOOL) {
        if (!message.thinks) {
            return;
        }
        const { toolCallId, content } = chunkMsg;
        for (let think of message.thinks) {
            if (think.type != THINK_TYPE.TOOL || !think.extra) {
                continue;
            }
            if (think.extra?.toolCallId == toolCallId) {
                think.thinkContent = appendToolCallResult(think.thinkContent, content);
                if (content.startsWith(TOOL_ERROR_PREFIX)) {
                    think.status = THINK_STATUS.ERROR;
                }
                return;
            }
        }
    }
};

const addReasoningContent = (
    msgIndex: number,
    content: {
        type?: string;
        title: string;
        reasoningContent: string;
        extra?: any;
    } = { title: '思考', reasoningContent: '' }
) => {
    const thinks = state.messages[msgIndex].thinks;
    const reasoningContent = content.reasoningContent;

    // 创建think对象的辅助函数
    const createThink = (status: string, codeId: number | string) => ({
        type: content.type,
        status,
        codeId: String(codeId),
        title: content.type == THINK_TYPE.TOOL ? '工具调用' : '思考',
        isCanExpand: true,
        isDefaultExpand: true,
        thinkTitle: content.title,
        thinkContent: reasoningContent,
        extra: content.extra,
    });

    // 如果没有thinks数组，初始化
    if (!thinks || thinks.length == 0) {
        state.messages[msgIndex].thinks = [createThink(THINK_STATUS.LOADING, 1)];
        return;
    }

    const thinkIndex = thinks.length - 1;
    const think = thinks[thinkIndex];

    // 如果title不同，结束当前think并创建新的
    if (think.title != content.title) {
        think.status = THINK_STATUS.SUCCESS;
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

    // 如果内容为空，直接返回
    if (!reasoningContent) {
        return;
    }

    // 否则创建新的think
    thinks.push(createThink(THINK_STATUS.LOADING, thinkIndex + 2));
};

const loadMessage = async () => {
    if (!props.sessionId) {
        return;
    }
    try {
        state.msgLoading = true;
        const messages = await aiApi.listMessages.request({ sessionKey: props.sessionId });

        const finalMessages = new Map<string, messageType>(); // messageId -> message
        const toolCallMessages = new Map<string, messageType[]>(); // messageId -> [message]
        const toolMessages = new Map<string, messageType>(); // toolCallId -> message
        // 第一轮遍历：分类消息
        for (let message of messages) {
            if (message.toolCalls && message.toolCalls.length > 0) {
                if (!toolCallMessages.has(message.messageId || message.time + '')) {
                    toolCallMessages.set(message.messageId || message.time + '', []);
                }
                toolCallMessages.get(message.messageId || message.time + '')?.push(message);
                continue;
            }
            if (message.role == ROLE.TOOL) {
                toolMessages.set(message.toolCallId || '', message);
                continue;
            }

            finalMessages.set(message.messageId || message.time + '', message);
        }

        // 第二轮遍历：处理工具调用和结果
        for (let [key, value] of toolCallMessages) {
            const message = finalMessages.get(key);
            if (!message) {
                continue;
            }
            const thinks = [];

            for (let toolCallMsg of value) {
                const tollCalls = toolCallMsg.toolCalls;
                if (!tollCalls) {
                    continue;
                }
                for (let toolCall of tollCalls) {
                    const toolResult = toolMessages.get(toolCall.id)?.content || '';
                    const think = {
                        type: THINK_TYPE.TOOL,
                        status: toolResult.startsWith(TOOL_ERROR_PREFIX) ? THINK_STATUS.ERROR : THINK_STATUS.SUCCESS,
                        codeId: String(toolCall.id),
                        title: '工具调用',
                        isCanExpand: true,
                        isDefaultExpand: false,
                        thinkTitle: formatToolThinkTitle(toolCall),
                        thinkContent: appendToolCallResult(JSON.stringify(toolCall), toolResult),
                    };
                    thinks.push(think);
                }
            }

            message.thinks = thinks;
        }

        state.messages = Array.from(finalMessages.values());
    } finally {
        state.msgLoading = false;
        scrollToBottom();
    }
};

const formatToolThinkTitle = (toolCall: ToolCall) => {
    return '工具调用 - ' + toolCall.function?.name || '';
};

const appendToolCallResult = (toolDesc: string, toolResult: string) => {
    return `${toolDesc}  工具调用结果: ${toolResult}`;
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
            placement: role === ROLE.AI ? 'start' : 'end',
            variant: role === ROLE.AI ? 'filled' : 'outlined', // 气泡的样式
            isFog: role === ROLE.AI, // AI 消息开启雾化效果
            maxWidth: BUBBLE_MAX_WIDTH,
        };
    }) as any[];
});

const sendUserMsg = (msg: string) => {
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
            sessionId: props.sessionId,
            content: msg,
        })
    );
};

const onSubmit = (value: SubmitResult) => {
    try {
        state.senderLoading = true;
        sendUserMsg(value.text);

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

// 记录进入思考中
let isThinking = false;

const handleThinkTagContent = (msgIndex: number, content?: string) => {
    if (!content) return;

    const hasThinkStart = content.includes('<think>');
    const hasThinkEnd = content.includes('</think>');

    if (hasThinkStart) isThinking = true;
    if (hasThinkEnd) isThinking = false;

    if (isThinking) {
        addReasoningContent(msgIndex, {
            title: '思考',
            reasoningContent: content.replace(/<think>|<\/think>/g, ''),
        });
    } else if (hasThinkEnd) {
        addReasoningContent(msgIndex);
    }
};
</script>

<style></style>
