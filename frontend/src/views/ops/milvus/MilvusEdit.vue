<template>
    <div>
        <el-drawer :title="title" v-model="dialogVisible" :before-close="onCancel" :destroy-on-close="true" :close-on-click-modal="false" size="40%">
            <template #header>
                <DrawerHeader :header="title" :back="onCancel" />
            </template>

            <el-form :model="form" ref="milvusFormRef" :rules="rules" label-width="auto">
                <el-form-item prop="tagCodePaths" :label="$t('tag.relateTag')" required>
                    <tag-tree-select multiple v-model="form.tagCodePaths" />
                </el-form-item>
                <el-form-item prop="name" :label="$t('common.name')" required>
                    <el-input v-model.trim="form.name" :placeholder="$t('common.pleaseInput')" auto-complete="off"></el-input>
                </el-form-item>
                <el-form-item prop="host" :label="$t('milvus.host')" required>
                    <el-input v-model.trim="form.host" :placeholder="$t('milvus.connAddress')" auto-complete="off" type="textarea"></el-input>
                </el-form-item>
                <el-form-item prop="username" :label="$t('common.username')">
                    <el-input v-model.trim="form.username"></el-input>
                </el-form-item>
                <el-form-item prop="password" :label="$t('common.password')">
                    <el-input type="password" show-password v-model.trim="form.password" autocomplete="new-password"> </el-input>
                </el-form-item>
                <el-form-item prop="database" :label="$t('milvus.database')">
                    <el-input v-model.trim="form.database" :placeholder="$t('milvus.dbNamePlaceholder')"></el-input>
                </el-form-item>
                <el-form-item prop="sshTunnelMachineId" :label="$t('machine.sshTunnel')">
                    <ssh-tunnel-select v-model="form.sshTunnelMachineId" />
                </el-form-item>
            </el-form>

            <template #footer>
                <div class="dialog-footer">
                    <el-button @click="onTestConn" :loading="testConnBtnLoading" type="success">{{ $t('ac.testConn') }}</el-button>
                    <el-button @click="onCancel()">{{ $t('common.cancel') }}</el-button>
                    <el-button type="primary" :loading="saveBtnLoading" @click="onConfirm">{{ $t('common.confirm') }}</el-button>
                </div>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { toRefs, reactive, watch, useTemplateRef } from 'vue';
import { milvusApi } from './api';
import { ElMessage } from 'element-plus';
import SshTunnelSelect from '../component/SshTunnelSelect.vue';
import DrawerHeader from '@/components/drawer-header/DrawerHeader.vue';
import { Rules } from '@/common/rule';
import TagTreeSelect from '@/views/ops/component/TagTreeSelect.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const props = defineProps({
    milvus: {
        type: [Boolean, Object],
    },
    title: {
        type: String,
    },
});

const dialogVisible = defineModel<boolean>('visible', { default: false });

const emit = defineEmits(['val-change', 'cancel']);

const rules = {
    code: [Rules.requiredInput('milvus.code')],
    name: [Rules.requiredInput('common.name')],
    host: [Rules.requiredInput('milvus.host')],
};

const milvusFormRef: any = useTemplateRef('milvusFormRef');

const state = reactive({
    form: {
        id: null,
        code: '',
        name: null,
        host: '',
        username: '',
        password: '',
        database: 'default',
        sshTunnelMachineId: null as any,
        tagCodePaths: [],
    },
    submitForm: {} as any,
});

const { form, submitForm } = toRefs(state);

const { isFetching: testConnBtnLoading, execute: testConnExec } = milvusApi.testConn.useApi(submitForm);
const { isFetching: saveBtnLoading, execute: saveMilvusExec } = milvusApi.save.useApi(submitForm);

watch(dialogVisible, () => {
    if (!dialogVisible.value) {
        return;
    }

    const milvus: any = props.milvus;
    if (milvus) {
        state.form = { ...milvus };
        state.form.tagCodePaths = milvus.tags.map((t: any) => t.codePath);
    } else {
        state.form = { database: 'default', sshTunnelMachineId: -1 } as any;
    }
});

const getReqForm = async () => {
    const reqForm = { ...state.form };
    if (!state.form.sshTunnelMachineId || state.form.sshTunnelMachineId <= 0) {
        reqForm.sshTunnelMachineId = -1;
    }
    return reqForm;
};

const onTestConn = async () => {
    await milvusFormRef.value?.validate();
    state.submitForm = await getReqForm();
    await testConnExec();
    ElMessage.success(t('milvus.connSuccess'));
};

const onConfirm = async () => {
    await milvusFormRef.value?.validate();
    state.submitForm = await getReqForm();
    await saveMilvusExec();
    ElMessage.success(t('milvus.savedSuccess'));
    emit('val-change', state.form);
    onCancel();
};

const onCancel = () => {
    dialogVisible.value = false;
    emit('cancel');
};
</script>

<style lang="scss" scoped></style>
