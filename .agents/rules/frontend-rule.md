---
trigger: always_on
---

你是一位资深的TypeScript前端工程师，严格遵循DRY/KISS原则，精通响应式设计模式，注重代码可维护性与可测试性，遵循Airbnb TypeScript代码规范，熟悉Vue等主流框架的最佳实践。

---

# Mayfly-Go 前端开发规范

## 核心技术栈

Vue 3 (Composition API) + TypeScript + Vite + Element Plus + Tailwind CSS + Pinia

---

## 综合开发示例

以下是一个完整的列表页 + 编辑对话框示例，涵盖了项目的所有核心开发规范：

### 枚举定义 (`src/views/system/enums.ts`)

```typescript
import { EnumValue } from '@/common/Enum';

export const AccountStatusEnum = {
    Enable: EnumValue.of(1, 'system.account.statusEnable').tagTypeSuccess(),
    Disable: EnumValue.of(-1, 'system.account.statusDisable').tagTypeDanger(),
};
```

**国际化配置** (`src/i18n/zh-cn/system.ts`):

```typescript
export default {
    system: {
        account: {
            statusEnable: '启用',
            statusDisable: '禁用',
            editAccount: '编辑账号',
            addAccount: '新增账号',
        },
    },
};
```

**国际化文件组织规范**:

- **按模块拆分**: 在 `src/i18n/{lang}/` 目录下，每个业务模块对应一个独立文件
- **文件命名**: 使用模块名小写，如 `system.ts`, `ai.ts`, `db.ts`, `machine.ts`
- **命名空间**: 以模块名为根 key（如 `system`, `ai`, `ops`），子功能作为嵌套对象
- **自动加载**: 项目通过 `import.meta.glob` 自动加载所有语言文件，新增文件无需手动注册

**示例结构**:

```
src/i18n/
├── zh-cn/
│   ├── common.ts      # 通用文本（按钮、状态等）
│   ├── system.ts      # 系统管理模块
│   ├── ai.ts          # AI 助手模块
│   ├── db.ts          # 数据库模块
│   ├── machine.ts     # 机器管理模块
│   └── ...
└── en/
    ├── common.ts
    ├── system.ts
    ├── ai.ts
    └── ...
```

**使用方式**:

```vue
<template>
    <!-- 模板中使用 $t() -->
    <h1>{{ $t('system.account.name') }}</h1>
    <el-button>{{ $t('common.save') }}</el-button>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 脚本中使用 t()
const message = t('common.success');
const confirmMsg = t('system.account.deleteAccountConfirm', { name: '张三' });
</script>
```

**注意事项**:

- ✅ 所有展示文本必须通过 `$t()` 或 `t()` 获取，禁止硬编码
- ✅ 枚举的 label 必须是国际化 key（如 `'system.account.statusEnable'`）
- ✅ 新增模块时，在 `zh-cn/` 和 `en/` 下同时创建对应的语言文件
- ❌ 禁止在组件中直接写中文/英文文本

### API 定义 (`src/views/system/api.ts`)

```typescript
import Api from '@/common/Api';

export const accountApi = {
    list: Api.newGet('/sys/accounts'),
    save: Api.newPost('/sys/accounts'),
    update: Api.newPut('/sys/accounts/{id}'),
    del: Api.newDelete('/sys/accounts/{id}'),
    changeStatus: Api.newPut('/sys/accounts/change-status/{id}/{status}'),
};
```

### 列表页组件 (`src/views/system/account/AccountList.vue`)

```vue
<template>
    <page-table ref="pageTableRef" :page-api="accountApi.list" :search-items="searchItems" v-model:query-form="query" :columns="columns">
        <template #tableHeader>
            <el-button v-auth="'account:add'" type="primary" @click="onAdd">
                {{ $t('common.create') }}
            </el-button>
            <el-button v-auth="'account:del'" type="danger" :disabled="selectionData.length === 0" @click="onBatchDelete">
                {{ $t('common.delete') }}
            </el-button>
        </template>

        <template #action="{ data }">
            <el-button link v-auth="'account:edit'" @click="onEdit(data)">
                {{ $t('common.edit') }}
            </el-button>
            <el-button link v-auth="'account:changeStatus'" :type="data.status === 1 ? 'danger' : 'success'" @click="onChangeStatus(data)">
                {{ data.status === 1 ? $t('common.disable') : $t('common.enable') }}
            </el-button>
            <el-button link v-auth="'account:del'" type="danger" @click="onDelete(data)">
                {{ $t('common.delete') }}
            </el-button>
        </template>
    </page-table>

    <AccountEdit v-model:visible="editVisible" :data="editData" @success="onEditSuccess" />
</template>

<script lang="ts" setup>
// ========== Imports ==========
import { ref } from 'vue';
import { accountApi } from '../api';
import { AccountStatusEnum } from '../enums';
import PageTable from '@/components/pagetable/PageTable.vue';
import { SearchItem, TableColumn } from '@/components/pagetable';
import { useI18nDeleteConfirm, useI18nDeleteSuccessMsg, useI18nOperateSuccessMsg } from '@/hooks/useI18n';
import AccountEdit from './AccountEdit.vue';

// ========== 常量定义 (as const) ==========
const PERMS = {
    ADD: 'account:add',
    EDIT: 'account:edit',
    DEL: 'account:del',
    CHANGE_STATUS: 'account:changeStatus',
} as const;

// ========== 类型定义 ==========
interface AccountInfo {
    id: number;
    username: string;
    name: string;
    status: number;
    createTime: string;
}

// ========== 响应式数据 ==========
const pageTableRef = ref();
const selectionData = ref<AccountInfo[]>([]);
const editVisible = ref(false);
const editData = ref<AccountInfo | null>(null);

const query = ref({
    username: '',
    status: null as number | null,
    pageNum: 1,
    pageSize: 10,
});

// ========== 计算属性 ==========
const searchItems = [SearchItem.input('username', 'common.username'), SearchItem.select('status', 'common.status', AccountStatusEnum)];

const columns = [
    TableColumn.new('username', 'common.username'),
    TableColumn.new('name', 'system.account.name'),
    TableColumn.new('status', 'common.status').typeTag(AccountStatusEnum),
    TableColumn.new('createTime', 'common.createTime').isTime(),
    TableColumn.new('action', 'common.operation').isSlot().fixedRight(),
];

// ========== 事件处理方法 (on 开头) ==========

/**
 * 新增账号
 */
const onAdd = () => {
    editData.value = null;
    editVisible.value = true;
};

/**
 * 编辑账号
 */
const onEdit = (row: AccountInfo) => {
    editData.value = row;
    editVisible.value = true;
};

/**
 * 删除单个账号
 */
const onDelete = async (row: AccountInfo) => {
    await useI18nDeleteConfirm(row.username);
    await accountApi.del.request({ id: row.id });
    useI18nDeleteSuccessMsg();
    pageTableRef.value?.search();
};

/**
 * 批量删除账号
 */
const onBatchDelete = async () => {
    const names = selectionData.value.map((item) => item.username).join('、');
    await useI18nDeleteConfirm(names);
    const ids = selectionData.value.map((item) => item.id).join(',');
    await accountApi.del.request({ id: ids });
    useI18nDeleteSuccessMsg();
    pageTableRef.value?.search();
};

/**
 * 修改账号状态
 */
const onChangeStatus = async (row: AccountInfo) => {
    const newStatus = row.status === 1 ? -1 : 1;
    await accountApi.changeStatus.request({
        id: row.id,
        status: newStatus,
    });
    useI18nOperateSuccessMsg();
    pageTableRef.value?.search();
};

/**
 * 编辑成功回调
 */
const onEditSuccess = () => {
    editVisible.value = false;
    pageTableRef.value?.search();
};
</script>
```

### 编辑对话框组件 (`src/views/system/account/AccountEdit.vue`)

```vue
<template>
    <el-dialog v-model="visible" :title="dialogTitle" width="500px" @close="onDialogClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
            <el-form-item :label="$t('common.username')" prop="username">
                <el-input v-model="form.username" :placeholder="$t('common.inputPlaceholder')" :disabled="!!form.id" />
            </el-form-item>

            <el-form-item :label="$t('system.account.name')" prop="name">
                <el-input v-model="form.name" />
            </el-form-item>

            <el-form-item :label="$t('common.status')" prop="status">
                <EnumSelect v-model="form.status" :enums="AccountStatusEnum" />
            </el-form-item>
        </el-form>

        <template #footer>
            <el-button @click="onCancel">
                {{ $t('common.cancel') }}
            </el-button>
            <el-button type="primary" :loading="submitting" @click="onSubmit">
                {{ $t('common.confirm') }}
            </el-button>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
// ========== Imports ==========
import { ref, reactive, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { accountApi } from '../api';
import { AccountStatusEnum } from '../enums';
import EnumSelect from '@/components/enumselect/EnumSelect.vue';
import { useI18nOperateSuccessMsg } from '@/hooks/useI18n';

// ========== Props & Emits ==========
interface Props {
    visible?: boolean;
    data?: AccountInfo | null;
}

const props = withDefaults(defineProps<Props>(), {
    visible: false,
    data: null,
});

const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
}>();

// ========== 类型定义 ==========
interface AccountInfo {
    id?: number;
    username: string;
    name: string;
    status: number;
}

// ========== 响应式数据 ==========
const formRef = ref();
const submitting = ref(false);

const form = reactive<AccountInfo>({
    id: undefined,
    username: '',
    name: '',
    status: 1,
});

// ========== 计算属性 ==========
const visible = computed({
    get: () => props.visible,
    set: (val) => emit('update:visible', val),
});

const { t } = useI18n();
const dialogTitle = computed(() => (form.id ? t('system.account.editAccount') : t('system.account.addAccount')));

const rules = {
    username: [
        { required: true, message: '请输入用户名', trigger: 'blur' },
        { min: 5, max: 16, message: '长度为5-16位', trigger: 'blur' },
    ],
    name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }],
};

// ========== 监听器 ==========
watch(
    () => props.data,
    (newVal) => {
        if (newVal) {
            Object.assign(form, newVal);
        } else {
            resetForm();
        }
    },
    { immediate: true }
);

// ========== 工具函数 ==========
const resetForm = () => {
    form.id = undefined;
    form.username = '';
    form.name = '';
    form.status = 1;
    formRef.value?.clearValidate();
};

// ========== 事件处理方法 ==========

/**
 * 提交表单
 */
const onSubmit = async () => {
    await formRef.value?.validate();

    submitting.value = true;
    try {
        if (form.id) {
            await accountApi.update.request(form);
        } else {
            await accountApi.save.request(form);
        }
        useI18nOperateSuccessMsg();
        visible.value = false;
        emit('success');
    } finally {
        submitting.value = false;
    }
};

/**
 * 取消操作
 */
const onCancel = () => {
    visible.value = false;
};

/**
 * 对话框关闭
 */
const onDialogClose = () => {
    resetForm();
};
</script>
```

---

## 核心规范总结

### ✅ 必须遵守

- **枚举使用**:
    - 在 `enums.ts` 中定义，使用 `EnumValue.of(value, label)`
    - 展示用 `<EnumTag>`，选择用 `<EnumSelect>`
    - 脚本中用 `EnumValue.getEnumByValue()` 获取枚举对象

- **API 调用**:
    - 在 `api.ts` 中定义，使用 `Api.newGet/Post/Put/Delete()`
    - 简单请求: `await api.xxx.request(params)`
    - 响应式: `const { execute, isFetching } = api.xxx.useApi()`
    - 表格集成: `<page-table :page-api="api.list" />`

- **命名规范**:
    - 事件方法必须以 `on` 开头: `onSubmit`, `onDelete`, `onEdit`
    - 变量/函数: camelCase
    - 常量: UPPER_SNAKE_CASE + `as const`
    - 组件: PascalCase
    - 文件: 组件用 PascalCase，其他用小写

- **代码组织顺序**:

    ```
    Imports
    Props/Emits
    常量定义 (as const)
    类型定义
    响应式数据
    计算属性
    监听器
    工具函数
    事件处理方法 (on 开头)
    ```

- **样式优先**: 使用 Tailwind CSS，支持 `dark:` 前缀

- **国际化规范**:
    - **文件组织**: 在 `src/i18n/{lang}/` 下按模块拆分文件（如 `system.ts`, `ai.ts`, `db.ts`）
    - **命名空间**: 以模块名为根 key，子功能嵌套（如 `system.account.name`）
    - **使用方式**: 模板用 `$t()`，脚本用 `t()`
    - **枚举关联**: 枚举 label 必须是国际化 key

- **权限控制**: 按钮用 `v-auth` 指令

- **类型安全**: 避免 `any`，使用可选链 `?.`

### ❌ 禁止事项

- ❌ 硬编码状态值，必须用枚举
- ❌ 使用 `Object.values().find()`，必须用 `EnumValue.getEnumByValue()`
- ❌ 直接调用 axios，必须通过 API 封装
- ❌ 文本硬编码，必须用国际化（`$t()` / `t()`）
- ❌ 事件方法不以 `on` 开头
- ❌ 使用固定高度计算，优先用 Flexbox
- ❌ 在组件中直接写中文/英文，必须配置到 i18n 文件

---

**最后更新**: 2026-04-19
