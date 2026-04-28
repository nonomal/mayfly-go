<template>
    <div class="flex h-full">
        <ConfigProvider :theme="isDark ? 'dark' : 'light'">
            <!-- 左侧会话列表 -->
            <aside class="flex flex-col border-r transition-all duration-300 ease-in-out" style="border-color: var(--el-border-color-light, #e4e7ed)">
                <Conversations
                    :active="currentSessionId"
                    :items="conversations"
                    @change="onChangeSession"
                    :label-max-width="200"
                    :show-tooltip="true"
                    row-key="key"
                    tooltip-placement="right"
                    :tooltip-offset="35"
                    show-to-top-btn
                    :menuTeleported="false"
                    show-built-in-menu
                    @menu-command="onMenuCommand"
                >
                    <template #header>
                        <div class="p-2 border-b" style="border-color: var(--el-border-color-lighter, #ebeef5)">
                            <el-button
                                icon="plus"
                                type="primary"
                                @click="createNewSession"
                                class="w-full shadow-sm hover:shadow-md transition-shadow duration-200"
                            >
                                {{ t('ai.assistant.newSession') }}
                            </el-button>
                        </div>
                    </template>

                    <!-- <template #menu="{ item }">
                    <div class="flex flex-col">
                        <el-button
                            v-for="menuItem in conversationMenuItems"
                            :key="menuItem.key"
                            link
                            size="small"
                            :icon="menuItem.icon"
                            @click.stop="onMenuCommand(menuItem.key, item)"
                        >
                            <span v-if="menuItem.label">{{ menuItem.label }}</span>
                        </el-button>
                    </div>
                </template> -->
                </Conversations>
            </aside>

            <!-- 右侧聊天区域 -->
            <main class="ml-3 flex-1 flex flex-col bg-linear-to-br from-gray-50 to-white dark:from-gray-800 dark:to-gray-900">
                <AiChat v-if="currentSessionId" :session-id="currentSessionId" @activate="onSessionCreated" />
                <div v-else class="flex-1 flex flex-col items-center justify-center text-gray-400 dark:text-gray-500 space-y-4">
                    <div class="text-6xl opacity-20">💬</div>
                    <p class="text-lg font-medium">{{ t('ai.assistant.startNewConversation') }}</p>
                    <p class="text-sm">{{ t('ai.assistant.selectOrCreateSession') }}</p>
                </div>
            </main>
        </ConfigProvider>
    </div>
</template>

<script setup lang="ts" name="AiAssistant">
import { notBlankI18n } from '@/common/assert';
import { formatDate } from '@/common/utils/format';
import { randomUuid } from '@/common/utils/string';
import { useI18nOperateSuccessMsg } from '@/hooks/useI18n';
import { ElMessageBox } from 'element-plus';
import { computed, defineAsyncComponent, onMounted, ref } from 'vue';
import { Conversations, ConfigProvider } from 'vue-element-plus-x';
import type { ConversationItem, ConversationMenuCommand } from 'vue-element-plus-x/types/Conversations';
import { useI18n } from 'vue-i18n';
import { aiApi } from './api';
import { useThemeConfig } from '@/store/themeConfig';

const AiChat = defineAsyncComponent(() => import('./AiChat.vue'));

const { t } = useI18n();

const themeConfig = useThemeConfig();
const isDark = computed(() => themeConfig.themeConfig.isDark);

const conversations = ref<ConversationItem[]>([]);
// 当前会话id（'-1' 表示新建会话）
const currentSessionId = ref<string>('');

/**
 * 加载会话列表
 */
const loadSessions = async () => {
    const sessions = await aiApi.listSessions.request();
    conversations.value = sessions.map((session) => {
        return {
            key: session.sessionKey,
            label: session.title,
            createTime: formatDate(session.createTime),
            updateTime: formatDate(session.updateTime),
        };
    });
    // 默认选中第一个会话
    if (!currentSessionId.value && sessions.length > 0) {
        switchSession(sessions[0].sessionKey);
    }
};

const onChangeSession = (item: ConversationItem) => {
    switchSession(item.key);
};

const createNewSession = async () => {
    currentSessionId.value = '-1';
};

const switchSession = (sessionId: string) => {
    currentSessionId.value = sessionId;
};

const onSessionCreated = async (sessionId: string) => {
    currentSessionId.value = sessionId;
    await loadSessions();
};

const deleteSession = async (sessionKey: string) => {
    await aiApi.deleteSession.request({ sessionKey });
    if (currentSessionId.value === sessionKey) {
        currentSessionId.value = '';
    }
    await loadSessions();
};

// 内置菜单点击方法
const onMenuCommand = async (command: ConversationMenuCommand, item: ConversationItem) => {
    if (command === 'delete') {
        deleteSession(item.key);
        return;
    }
    if (command === 'rename') {
        ElMessageBox.prompt('', t('common.name'), {
            confirmButtonText: t('common.confirm'),
            cancelButtonText: t('common.cancel'),
        }).then(async ({ value }) => {
            notBlankI18n(value, 'common.name');
            await aiApi.renameSession.request({ sessionKey: item.key, title: value });
            useI18nOperateSuccessMsg();
            loadSessions();
        });
    }
};

onMounted(() => {
    loadSessions();
});
</script>

<style></style>
