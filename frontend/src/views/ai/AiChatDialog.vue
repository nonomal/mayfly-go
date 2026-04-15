<template>
    <el-drawer
        :title="title"
        v-model="dialogVisible"
        z-index="2000"
        top="32px"
        size="70%"
        :close-on-click-modal="false"
        :show-close="true"
        @close="closeDialog"
    >
        <template #header>
            <DrawerHeader
                :header="title"
                :back="
                    () => {
                        dialogVisible = false;
                    }
                "
            />
        </template>

        <AiAssistant></AiAssistant>
        <!-- <AiChat ref="aiChatRef" :title="title" :show-close="false" /> -->
    </el-drawer>
</template>

<script setup lang="ts" name="AiChatDialog">
import { ref, onMounted, useTemplateRef, watch, defineAsyncComponent } from 'vue';

const DrawerHeader = defineAsyncComponent(() => import('@/components/drawer-header/DrawerHeader.vue'));
const AiAssistant = defineAsyncComponent(() => import('./AiAssistant.vue'));

const props = defineProps({
    title: {
        type: String,
        default: '',
    },
});

const emit = defineEmits(['close', 'send', 'refresh']);

const dialogVisible = defineModel<Boolean>('visible', { required: true });
// const aiChatRef = useTemplateRef<InstanceType<typeof AiChat>>('aiChatRef');

const handleSend = (message: string) => {
    emit('send', message);
};

const handleRefresh = () => {
    emit('refresh');
};

const closeDialog = () => {
    emit('close');
};
</script>
