<template>
    <el-tabs v-model="activeTab" type="border-card">
        <!-- 数据库管理 -->
        <el-tab-pane :label="$t('milvus.databaseManagement')" name="databases">
            <Databases :milvus-id="milvusId" v-if="activeTab === 'databases'" @use="onUseDb" />
        </el-tab-pane>

        <!-- Collection 管理 -->
        <el-tab-pane :label="$t('milvus.collectionManagement')" name="collections">
            <Collections :milvus-id="milvusId" v-if="activeTab === 'collections'" @change-tab="(name) => (activeTab = name)" />
        </el-tab-pane>

        <!-- 数据操作 -->
        <el-tab-pane :label="$t('milvus.dataOperation')" name="data">
            <DataOperation :milvus-id="milvusId" v-if="activeTab === 'data'" />
        </el-tab-pane>

        <!-- 分区管理 -->
        <el-tab-pane :label="$t('milvus.partitionManagement')" name="partitions">
            <Partitions :milvus-id="milvusId" v-if="activeTab === 'partitions'" />
        </el-tab-pane>
        <!-- 用户权限 -->
        <el-tab-pane :label="$t('milvus.userPermission')" name="users">
            <Users :milvus-id="milvusId" v-if="activeTab === 'users'" />
        </el-tab-pane>

        <!-- 角色管理 -->
        <el-tab-pane :label="$t('milvus.roleManagement')" name="roles">
            <Roles :milvus-id="milvusId" v-if="activeTab === 'roles'" />
        </el-tab-pane>

        <!-- 权限组管理 -->
        <!-- <el-tab-pane :label="$t('milvus.privilegeGroupManagement')" name="privilegeGroups">
            <PrivilegeGroups v-if="activeTab === 'privilegeGroups'" :milvus-id="milvusId" />
        </el-tab-pane> -->

        <!-- 资源组 -->
        <el-tab-pane :label="$t('milvus.resourceGroup')" name="resourceGroups">
            <ResourceGroups :milvus-id="milvusId" v-if="activeTab === 'resourceGroups'" />
        </el-tab-pane>

        <!-- 系统信息 -->
        <el-tab-pane :label="$t('milvus.systemInfo')" name="system">
            <SystemInfo :milvus-id="milvusId" v-if="activeTab === 'system'" />
        </el-tab-pane>
    </el-tabs>
</template>

<script setup lang="ts">
import { onMounted, ref, getCurrentInstance } from 'vue';
import Databases from '../components/Databases.vue';
import Collections from '../components/Collections.vue';
import DataOperation from '../components/DataOperation.vue';
import Partitions from '../components/Partitions.vue';
import Users from '../components/Users.vue';
import Roles from '../components/Roles.vue';
import PrivilegeGroups from '../components/PrivilegeGroups.vue';
import ResourceGroups from '../components/ResourceGroups.vue';
import SystemInfo from '../components/SystemInfo.vue';
import { MilvusOpComp } from '@/views/ops/milvus/resource/index';

const milvusId = ref<number>(0);

const emits = defineEmits(['init']);
const initMilvus = (params: any) => {
    activeTab.value = 'databases';
    milvusId.value = params.id;
};

const activeTab = ref('databases');

const onUseDb = (db: string) => {
    activeTab.value = 'collections';
};

onMounted(() => {
    emits('init', { name: MilvusOpComp.name, ref: getCurrentInstance()?.exposed });
});

defineExpose({
    initMilvus,
});
</script>

<style scoped></style>
