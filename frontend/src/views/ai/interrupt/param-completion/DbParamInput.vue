<template>
    <div class="db-param-input">
        <div class="text-xs text-gray-500 dark:text-gray-400 mb-2">
            {{ t('ai.interrupt.paramCompletion.selectDbHint') }}
        </div>

        <!-- 未选择时显示选择器 -->
        <DbSelectTree
            v-if="!dbValue.dbId"
            v-model:db-id="dbValue.dbId"
            v-model:inst-name="dbValue.instanceName"
            v-model:db-name="dbValue.dbName"
            v-model:tag-path="dbValue.tagPath"
            v-model:db-type="dbValue.dbType"
            @select-db="onSelectDb"
        />

        <!-- 已选中的数据库信息 -->
        <div v-if="dbValue.dbId" class="mt-3 p-3 bg-primary-50 dark:bg-primary-900/20 rounded border border-primary-200 dark:border-primary-800">
            <div class="flex items-center gap-2">
                <SvgIcon :name="getDbDialect(dbValue.dbType)?.getInfo().icon || 'DataLine'" :size="20" />
                <div class="flex-1">
                    <div class="font-medium text-sm">{{ dbValue.instanceName }} - {{ dbValue.dbName }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ t('ai.interrupt.paramCompletion.dbType') }}: {{ dbValue.dbType }}</div>
                </div>
                <el-icon v-if="isConfirmed" class="text-success">
                    <Check />
                </el-icon>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import DbSelectTree from '@/views/ops/db/component/DbSelectTree.vue';
import { getDbDialect } from '@/views/ops/db/dialect';

interface DbParamValue {
    dbId: number;
    dbName: string;
    dbType: string;
    instanceName: string;
    tagPath: string;
}

interface Props {
    params: any[];
    readonly?: boolean;
    isConfirmed?: boolean;
    modelValue?: DbParamValue;
}

const props = withDefaults(defineProps<Props>(), {
    readonly: false,
    isConfirmed: false,
    modelValue: () => ({
        dbId: 0,
        dbName: '',
        dbType: '',
        instanceName: '',
        tagPath: '',
    }),
});

const emit = defineEmits<{
    change: [value: DbParamValue];
}>();

const { t } = useI18n();

// 使用 defineModel 实现双向绑定
const dbValue = defineModel<DbParamValue>('modelValue', {
    default: () => ({
        dbId: 0,
        dbName: '',
        dbType: '',
        instanceName: '',
        tagPath: '',
    }),
});

// 处理数据库选择
const onSelectDb = (params: any) => {
    console.log('[DbParamInput] Database selected:', params);
};

// 检查是否有效
const isValid = () => {
    return dbValue.value.dbId > 0;
};

// 获取参数值
const getValues = () => {
    return {
        id: dbValue.value.dbId,
        params: { ...dbValue.value },
        displayName: `${dbValue.value.instanceName} - ${dbValue.value.dbName}`,
    };
};

// 获取需要缓存的参数名
const getCacheableParams = () => {
    return props.params.filter((p: any) => p.cacheable === true).map((p: any) => p.param);
};

defineExpose({
    isValid,
    getValues,
    getCacheableParams,
});
</script>

<style scoped>
.db-param-input {
    padding: 0.5rem;
}
</style>
