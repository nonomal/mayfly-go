<template>
    <div>
        <el-drawer :title="title" v-model="dialogVisible" :before-close="onCancel" :close-on-click-modal="false" size="40%" :destroy-on-close="true">
            <el-form :model="form" ref="kafkaFormRef" :rules="rules" label-width="auto">
                <el-form-item prop="tagCodePaths" :label="$t('tag.relateTag')" required>
                    <tag-tree-select multiple v-model="form.tagCodePaths" />
                </el-form-item>

                <el-form-item prop="name" :label="$t('common.name')" required>
                    <el-input v-model.trim="form.name" auto-complete="off"></el-input>
                </el-form-item>
                <el-form-item prop="hosts" label="Hosts" required>
                    <el-input
                        type="textarea"
                        :rows="2"
                        v-model.trim="form.hosts"
                        placeholder="Kafka 连接地址，格式: host1:port1,host2:port2 或单个 broker"
                        auto-complete="off"
                    />
                </el-form-item>
                <el-form-item prop="saslMechanism" :label="$t('mq.kafka.sasl_mechanism')">
                    <el-select
                        v-model="form.saslMechanism"
                        :options="sasl_mechanism_options"
                        :placeholder="t('mq.kafka.sasl_mechanism_placeholder')"
                        filterable
                        clearable
                    />
                </el-form-item>
                <el-form-item prop="username" :label="$t('mq.kafka.username')">
                    <el-input v-model.trim="form.username" auto-complete="off"></el-input>
                </el-form-item>
                <el-form-item prop="password" :label="$t('common.password')">
                    <el-input v-model.trim="form.password" auto-complete="off"></el-input>
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
import { toRefs, reactive, watchEffect, useTemplateRef } from 'vue';
import { ElMessage } from 'element-plus';
import TagTreeSelect from '../../component/TagTreeSelect.vue';
import SshTunnelSelect from '../../component/SshTunnelSelect.vue';
import { useI18nFormValidate, useI18nSaveSuccessMsg } from '@/hooks/useI18n';
import { useI18n } from 'vue-i18n';
import { Rules } from '@/common/rule';
import { mqApi } from '@/views/ops/mq/api';

const { t } = useI18n();

const props = defineProps({
    kafka: {
        type: [Boolean, Object],
    },
    title: {
        type: String,
    },
});

const sasl_mechanism_options = [
    {
        label: 'PLAIN',
        value: 'PLAIN',
    },
    {
        label: 'SCRAM-SHA-256',
        value: 'SCRAM-SHA-256',
    },
    {
        label: 'SCRAM-SHA-512',
        value: 'SCRAM-SHA-512',
    },
];

const dialogVisible = defineModel<boolean>('visible', { default: false });

//定义事件
const emit = defineEmits(['cancel', 'val-change']);

const rules = {
    tagCodePaths: [Rules.requiredSelect('tag.relateTag')],
    name: [Rules.requiredInput('common.name')],
    uri: [Rules.requiredInput('kafka.connUrl')],
};

const kafkaFormRef: any = useTemplateRef('kafkaFormRef');

const state = reactive({
    tabActiveName: 'basic',
    form: {
        id: null,
        code: '',
        name: null,
        hosts: '',
        username: '',
        saslMechanism: 'PLAIN',
        password: '',
        sshTunnelMachineId: null as any,
        tagCodePaths: [],
    },
});

const { tabActiveName, form } = toRefs(state);

const { isFetching: testConnBtnLoading, execute: testConnExec } = mqApi.KafkaTestConn.useApi();
const { isFetching: saveBtnLoading, execute: saveKafkaExec } = mqApi.kafkaSave.useApi();

watchEffect(() => {
    if (!dialogVisible.value) {
        return;
    }
    state.tabActiveName = 'basic';
    const kafka: any = props.kafka;
    if (kafka) {
        state.form = { ...kafka };
        state.form.tagCodePaths = kafka.tags.map((t: any) => t.codePath);
    } else {
        state.form = { db: 0, tagCodePaths: [] } as any;
    }
});

const getReqForm = () => {
    const reqForm = { ...state.form };
    if (!state.form.sshTunnelMachineId || state.form.sshTunnelMachineId <= 0) {
        reqForm.sshTunnelMachineId = -1;
    }
    return reqForm;
};

const onTestConn = async () => {
    await useI18nFormValidate(kafkaFormRef);
    await testConnExec(getReqForm());
    ElMessage.success(t('ac.connSuccess'));
};

const onConfirm = async () => {
    await useI18nFormValidate(kafkaFormRef);
    await saveKafkaExec(getReqForm());
    useI18nSaveSuccessMsg();
    emit('val-change', state.form);
    onCancel();
};

const onCancel = () => {
    dialogVisible.value = false;
    emit('cancel');
};
</script>
<style lang="scss"></style>
