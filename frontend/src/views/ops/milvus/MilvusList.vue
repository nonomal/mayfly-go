<template>
    <div class="h-full">
        <page-table
            ref="pageTableRef"
            :page-api="milvusApi.list"
            :searchItems="searchItems"
            v-model:query-form="query"
            :show-selection="true"
            v-model:selection-data="selectionData"
            :columns="columns"
            lazy
        >
            <template #tableHeader>
                <el-button v-auth="perms.inst_save" type="primary" icon="plus" @click="editMilvus()" plain>{{ $t('common.create') }}</el-button>
                <el-button v-auth="perms.inst_del" type="danger" icon="delete" :disabled="selectionData.length < 1" @click="deleteMilvus" plain>
                    {{ $t('common.delete') }}
                </el-button>
            </template>

            <template #tagPath="{ data }">
                <resource-tags :tags="data.tags" />
            </template>

            <template #action="{ data }">
                <el-button v-auth="perms.inst_save" @click="editMilvus(data)" link type="primary">{{ $t('common.edit') }}</el-button>
            </template>
        </page-table>

        <milvus-edit @val-change="search()" :title="milvusEditDialog.title" v-model:visible="milvusEditDialog.visible" v-model:milvus="milvusEditDialog.data" />
    </div>
</template>

<script setup lang="ts">
import { defineAsyncComponent, ref, Ref } from 'vue';
import { ElMessage } from 'element-plus';
import PageTable from '@/components/pagetable/PageTable.vue';
import { milvusApi, perms } from './api';
import type { IMilvus } from './types';
import { TableColumn } from '@/components/pagetable';
import { useI18nCreateTitle, useI18nDeleteConfirm, useI18nDeleteSuccessMsg, useI18nEditTitle } from '@/hooks/useI18n';
import { SearchItem } from '@/components/pagetable/SearchForm';
import { getTagPathSearchItem } from '@/views/ops/component/tag';
import { TagResourceTypePath } from '@/common/commonEnum';
import ResourceTags from '../component/ResourceTags.vue';

const MilvusEdit = defineAsyncComponent(() => import('./MilvusEdit.vue'));

const pageTableRef: Ref<any> = ref(null);

const query = ref({
    pageNum: 1,
    pageSize: 0,
});

const selectionData = ref([]);

const searchItems = [SearchItem.input('keyword', 'common.keyword').withPlaceholder('db.keywordPlaceholder'), getTagPathSearchItem(TagResourceTypePath.Db)];

const columns = ref([
    TableColumn.new('tags[0].tagPath', 'tag.relateTag').isSlot('tagPath').setAddWidth(20),
    TableColumn.new('name', 'common.name').setMinWidth(150),
    TableColumn.new('host', 'milvus.host').setMinWidth(200),
    TableColumn.new('username', 'mq.kafka.username'),
    TableColumn.new('password', 'common.password'),
    TableColumn.new('createTime', 'common.createTime').setMinWidth(180),
    TableColumn.new('creator', 'common.creator'),
    TableColumn.new('code', 'Code').setMinWidth(150),
    TableColumn.new('action', 'common.operation').isSlot().setMinWidth(100).fixedRight().alignCenter(),
]);

const milvusEditDialog = ref({
    title: '',
    visible: false,
    data: null as any,
});

const editMilvus = (data?: IMilvus) => {
    milvusEditDialog.value = {
        title: data ? useI18nEditTitle('Milvus') : useI18nCreateTitle('Milvus'),
        visible: true,
        data: data || null,
    };
};

const deleteMilvus = async () => {
    const records = selectionData.value || [];
    if (records.length === 0) {
        ElMessage.warning('请选择要删除的数据');
        return;
    }
    const ids = records.map((r: any) => r.id).join(',');

    await useI18nDeleteConfirm('Milvus: ' + ids);
    milvusApi.delete.request(ids).then(() => {
        useI18nDeleteSuccessMsg();
        search();
    });
};

const search = () => {
    pageTableRef.value?.search();
};

defineExpose({ search });
</script>
