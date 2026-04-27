<template>
    <el-space>
        <el-select size="small" v-model="selectedCollection" style="min-width: 200px" @change="loadList" filterable clearable>
            <el-option v-for="item in collections" :key="item" :label="item" :value="item" />
        </el-select>

        <el-button type="primary" size="small" icon="plus" @click="handleCreate">
            {{ $t('milvus.createPartition') }}
        </el-button>
        <el-button size="small" text icon="refresh" @click="loadList" :loading="loading" />
    </el-space>

    <el-table :data="list">
        <el-table-column prop="id" label="id" />
        <el-table-column prop="name" :label="$t('milvus.partitionName')" />
        <el-table-column prop="createTime" :label="$t('common.createTime')" />
        <el-table-column :label="$t('common.operation')" width="200">
            <template #default="{ row }">
                <el-button size="small" type="warning" plain @click="handleRelease(row)">{{ $t('milvus.release') }}</el-button>
                <el-button size="small" type="danger" @click="handleDrop(row)">{{ $t('common.delete') }}</el-button>
            </template>
        </el-table-column>
    </el-table>

    <el-dialog v-model="createDialog.visible" :title="$t('milvus.createPartition')" width="500px">
        <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="auto">
            <el-form-item :label="$t('milvus.partitionName')" prop="name">
                <el-input v-model="createForm.name" :placeholder="$t('milvus.partitionNamePlaceholder')"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="createDialog.visible = false">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="submitCreate" :loading="createLoading">{{ $t('common.confirm') }}</el-button>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { ElMessage, FormInstance } from 'element-plus';
import { milvusApi } from '../api';
import { Rules } from '@/common/rule';
import { useI18n } from 'vue-i18n';
import { useI18nConfirm } from '@/hooks/useI18n';
import { useMilvusStore } from '@/views/ops/milvus/resource/store';
import { storeToRefs } from 'pinia';

const milvusStore = useMilvusStore();
const { collections, selectedCollection } = storeToRefs(milvusStore);

const { t } = useI18n();
const props = defineProps<{
    milvusId: number;
}>();

const list = ref<any[]>([]);
const createDialog = ref({
    visible: false,
});
const createFormRef = ref<FormInstance>();
const loading = ref(false);
const createLoading = ref(false);
const createForm = ref({
    name: '',
});

const createRules = {
    name: [Rules.requiredInput('milvus.partitionName')],
};

const loadList = async () => {
    loading.value = true;
    // 需要先选择 collection，这里暂时加载所有分区
    try {
        const res = await milvusApi.listPartitions(props.milvusId, milvusStore.selectedCollection);
        list.value = res || [];
    } finally {
        loading.value = false;
    }
};

const handleCreate = () => {
    createForm.value = { name: '' };
    createDialog.value.visible = true;
};

const submitCreate = async () => {
    if (!createFormRef.value) return;

    await createFormRef.value.validate(async (valid) => {
        if (!valid) return;

        createLoading.value = true;
        try {
            await milvusApi.createPartition(props.milvusId, milvusStore.selectedCollection, createForm.value);
            ElMessage.success(t('milvus.createdSuccess'));
            createDialog.value.visible = false;
            await loadList();
        } finally {
            createLoading.value = false;
        }
    });
};

const handleDrop = async (row: any) => {
    await useI18nConfirm('milvus.confirmDeletePartition', { name: row.name });
    await milvusApi.dropPartition(props.milvusId, milvusStore.selectedCollection, row.name);
    ElMessage.success(t('milvus.deletedSuccess'));
    await loadList();
};

const handleRelease = async (row: any) => {
    await milvusApi.releasePartition(props.milvusId, milvusStore.selectedCollection, row.name);
    ElMessage.success(t('milvus.releasedSuccess'));
};

onMounted(() => {
    loadList();
});

watch(
    () => props.milvusId,
    () => {
        list.value = [];
        loadList();
        milvusStore.clear();
    }
);
</script>

<style scoped></style>
