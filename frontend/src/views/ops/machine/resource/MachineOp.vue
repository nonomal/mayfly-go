<template>
    <div class="h-full machine-terminal-tabs">
        <el-tabs v-if="state.tabs.size > 0" type="card" @tab-remove="onRemoveTab" v-model="state.activeTermName" class="!h-full w-full">
            <el-tab-pane class="h-full! flex flex-col" closable v-for="dt in state.tabs.values()" :label="dt.label" :name="dt.key" :key="dt.key">
                <template #label>
                    <el-popconfirm @confirm="handleReconnect(dt, true)" :title="$t('machine.reConnTips')" v-if="dt.type === 'terminal'">
                        <template #reference>
                            <el-icon
                                class="mr-1"
                                :color="EnumValue.getEnumByValue(TerminalStatusEnum, dt.status)?.extra?.iconColor"
                                :title="dt.status == TerminalStatusEnum.Connected.value ? '' : $t('machine.clickReConn')"
                            >
                                <Connection />
                            </el-icon>
                        </template>
                    </el-popconfirm>
                    <el-popover :show-after="1000" placement="bottom-start" trigger="hover" :width="250">
                        <template #reference>
                            <div>
                                <span class="machine-terminal-tab-label">{{ dt.label }}</span>
                            </div>
                        </template>
                        <template #default>
                            <el-descriptions :column="1" size="small">
                                <el-descriptions-item :label="$t('common.name')"> {{ dt.params?.name }} </el-descriptions-item>
                                <el-descriptions-item label="host"> {{ dt.params?.ip }} : {{ dt.params?.port }} </el-descriptions-item>
                                <el-descriptions-item label="username"> {{ dt.params?.selectAuthCert.username }} </el-descriptions-item>
                                <el-descriptions-item label="remark"> {{ dt.params?.remark }} </el-descriptions-item>
                            </el-descriptions>
                        </template>
                    </el-popover>
                </template>

                <!-- 终端类型 tab -->
                <div v-if="dt.type === 'terminal'" class="terminal-wrapper flex-1 min-h-0">
                    <TerminalBody
                        v-if="dt.params.protocol == MachineProtocolEnum.Ssh.value"
                        :mount-init="false"
                        @status-change="terminalStatusChange(dt.key, $event)"
                        :ref="(el: any) => setTerminalRef(el, dt.key)"
                        :socket-url="dt.socketUrl"
                        :machine-id="dt.params.id"
                        :auth-cert-name="dt.authCert"
                        :file-id="0"
                        :protocol="dt.params.protocol"
                    />
                    <machine-rdp
                        v-if="dt.params.protocol != MachineProtocolEnum.Ssh.value"
                        :machine-id="dt.params.id"
                        :auth-cert="dt.authCert"
                        :protocol="dt.params.protocol"
                        :ref="(el: any) => setTerminalRef(el, dt.key)"
                        @status-change="terminalStatusChange(dt.key, $event)"
                    />
                </div>

                <!-- 文件操作类型 tab -->
                <div v-if="dt.type === 'file'" class="file-wrapper flex-1 min-h-0">
                    <machine-file :machine-id="dt.machineId" :auth-cert-name="dt.authCertName" :protocol="dt.protocol" :file-id="dt.fileId" :path="dt.path" />
                </div>
            </el-tab-pane>
        </el-tabs>

        <el-dialog v-if="infoDialog.visible" v-model="infoDialog.visible">
            <el-descriptions :title="$t('common.detail')" :column="3" border>
                <el-descriptions-item :span="1.5" label="ID">{{ infoDialog.data.id }}</el-descriptions-item>
                <el-descriptions-item :span="1.5" :label="$t('common.name')">{{ infoDialog.data.name }}</el-descriptions-item>

                <el-descriptions-item :span="3" :label="$t('tag.relateTag')">
                    <TagCodePath :path="infoDialog.data.tags" />
                </el-descriptions-item>

                <el-descriptions-item :span="2" label="IP">{{ infoDialog.data.ip }}</el-descriptions-item>
                <el-descriptions-item :span="1" :label="$t('machine.port')">{{ infoDialog.data.port }}</el-descriptions-item>

                <el-descriptions-item :span="3" :label="$t('common.remark')">{{ infoDialog.data.remark }}</el-descriptions-item>

                <el-descriptions-item :span="1.5" :label="$t('machine.sshTunnel')"
                    >{{ infoDialog.data.sshTunnelMachineId > 0 ? $t('common.yes') : $t('common.no') }}
                </el-descriptions-item>
                <el-descriptions-item :span="1.5" :label="$t('machine.terminalPlayback')"
                    >{{ infoDialog.data.enableRecorder == 1 ? $t('common.yes') : $t('common.no') }}
                </el-descriptions-item>

                <el-descriptions-item :span="2" :label="$t('common.createTime')">
                    {{ formatDate(infoDialog.data.createTime) }}
                </el-descriptions-item>
                <el-descriptions-item :span="1" :label="$t('common.creator')">
                    {{ infoDialog.data.creator }}
                </el-descriptions-item>

                <el-descriptions-item :span="2" :label="$t('common.updateTime')">
                    {{ formatDate(infoDialog.data.updateTime) }}
                </el-descriptions-item>
                <el-descriptions-item :span="1" :label="$t('common.modifier')">
                    {{ infoDialog.data.modifier }}
                </el-descriptions-item>
            </el-descriptions>
        </el-dialog>

        <process-list v-model:visible="processDialog.visible" v-model:machineId="processDialog.machineId" />

        <script-manage
            :title="serviceDialog.title"
            v-model:visible="serviceDialog.visible"
            v-model:machineId="serviceDialog.machineId"
            :auth-cert-name="serviceDialog.authCertName"
        />

        <file-conf-list
            v-model:visible="fileDialog.visible"
            :machine-id="fileDialog.machine?.id"
            :auth-cert-name="fileDialog.machine?.selectAuthCert?.name"
            :protocol="fileDialog.machine?.protocol"
            :open-file-manager="false"
            @select="onFileConfigSelect"
        />

        <machine-stats v-model:visible="machineStatsDialog.visible" :machineId="machineStatsDialog.machineId" :title="machineStatsDialog.title" />

        <machine-rec v-model:visible="machineRecDialog.visible" :machineId="machineRecDialog.machineId" :title="machineRecDialog.title" />
    </div>
</template>

<script lang="ts" setup>
import EnumValue from '@/common/Enum';
import { formatDate } from '@/common/utils/format';
import { hasPerms } from '@/components/auth/auth';
import MachineRdp from '@/components/terminal-rdp/MachineRdp.vue';
import TerminalBody from '@/components/terminal/TerminalBody.vue';
import { TerminalStatus, TerminalStatusEnum } from '@/components/terminal/common';
import { ResourceOpCtx } from '@/views/ops/component/tag';
import MachineFile from '@/views/ops/machine/file/MachineFile.vue';
import { MachineOpComp } from '@/views/ops/machine/resource';
import { ResourceOpCtxKey } from '@/views/ops/resource/resource';
import { defineAsyncComponent, getCurrentInstance, inject, nextTick, onMounted, reactive, toRefs, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import TagCodePath from '../../component/TagCodePath.vue';
import { getMachineTerminalSocketUrl } from '../api';
import { MachineProtocolEnum } from '../enums';

// 组件
const ScriptManage = defineAsyncComponent(() => import('../ScriptManage.vue'));
const FileConfList = defineAsyncComponent(() => import('../file/FileConfList.vue'));
const MachineStats = defineAsyncComponent(() => import('../MachineStats.vue'));
const MachineRec = defineAsyncComponent(() => import('../MachineRec.vue'));
const ProcessList = defineAsyncComponent(() => import('../ProcessList.vue'));

const { t } = useI18n();

const router = useRouter();

// 机器信息类型定义
interface MachineInfo {
    id: number;
    name: string;
    ip: string;
    port: number;
    protocol: number;
    remark?: string;
    selectAuthCert: {
        name: string;
        username: string;
    };
}

const perms = {
    addMachine: 'machine:add',
    updateMachine: 'machine:update',
    delMachine: 'machine:del',
    terminal: 'machine:terminal',
    closeCli: 'machine:close-cli',
};

// 该用户拥有的的操作列按钮权限，使用v-if进行判断，v-auth对el-dropdown-item无效
const actionBtns = hasPerms([perms.updateMachine, perms.closeCli]);

const emits = defineEmits(['init']);

const resourceOpCtx: ResourceOpCtx | undefined = inject(ResourceOpCtxKey);

const state = reactive({
    defaultExpendKey: [] as any,
    params: {
        pageNum: 1,
        pageSize: 0,
        ip: null,
        name: null,
        tagPath: '',
    },
    infoDialog: {
        visible: false,
        data: null as any,
    },
    serviceDialog: {
        visible: false,
        machineId: 0,
        authCertName: '',
        title: '',
    },
    processDialog: {
        visible: false,
        machineId: 0,
    },
    fileDialog: {
        visible: false,
        machine: null as MachineInfo | null,
    },
    machineStatsDialog: {
        visible: false,
        stats: null,
        title: '',
        machineId: 0,
    },
    machineRecDialog: {
        visible: false,
        machineId: 0,
        title: '',
    },
    activeTermName: '',
    tabs: new Map<string, any>(),
});

const { infoDialog, serviceDialog, processDialog, fileDialog, machineStatsDialog, machineRecDialog } = toRefs(state);

let openIds: any = {};

watch(
    () => state.activeTermName,
    (newValue, oldValue) => {
        fitTerminal();

        // 只有终端类型才需要 blur/focus
        const oldTab = state.tabs.get(oldValue);
        const newTab = state.tabs.get(newValue);

        if (oldTab?.type === 'terminal') {
            terminalRefs[oldValue]?.blur && terminalRefs[oldValue]?.blur();
        }
        if (newTab?.type === 'terminal') {
            terminalRefs[newValue]?.focus && terminalRefs[newValue]?.focus();
        }

        resourceOpCtx?.setCurrentTreeKey(newTab?.authCert || newTab?.authCertName);
    }
);

onMounted(() => {
    emits('init', { name: MachineOpComp.name, ref: getCurrentInstance()?.exposed });
});

const openTerminal = (machine: any, ex?: boolean) => {
    // 授权凭证名
    const ac = machine.selectAuthCert.name;

    // 新窗口打开
    if (ex) {
        if (machine.protocol == MachineProtocolEnum.Ssh.value) {
            const { href } = router.resolve({
                path: `/machine/terminal`,
                query: {
                    ac,
                    name: machine.name,
                },
            });
            window.open(href, '_blank');
            return;
        }
        if (machine.protocol == MachineProtocolEnum.Rdp.value) {
            const { href } = router.resolve({
                path: `/machine/terminal-rdp`,
                query: {
                    machineId: machine.id,
                    ac: ac,
                    name: machine.name,
                },
            });
            window.open(href, '_blank');
            return;
        }
    }

    let { name } = machine;
    const labelName = `${machine.selectAuthCert.username}@${name}`;

    // 同一个机器的终端打开多次，key后添加下划线和数字区分
    openIds[ac] = openIds[ac] ? ++openIds[ac] : 1;
    let sameIndex = openIds[ac];

    let key = `${ac}_${sameIndex}`;
    // 只保留name的15个字，超出部分只保留前后10个字符，中间用省略号代替
    const label = labelName.length > 15 ? labelName.slice(0, 10) + '...' + labelName.slice(-10) : labelName;

    let tab = {
        key,
        label: `${label}${sameIndex === 1 ? '' : ':' + sameIndex}`, // label组成为:总打开term次数+name+同一个机器打开的次数
        type: 'terminal',
        params: machine,
        authCert: ac,
        socketUrl: getMachineTerminalSocketUrl(ac),
        status: TerminalStatusEnum.Disconnected.value,
    };

    state.tabs.set(key, tab);

    nextTick(() => {
        handleReconnect(tab);
        state.activeTermName = key;
        setTimeout(() => fitTerminal(), 300);
    });
};

const serviceManager = (row: any) => {
    const authCert = row.selectAuthCert;
    state.serviceDialog.machineId = row.id;
    state.serviceDialog.visible = true;
    state.serviceDialog.authCertName = authCert.name;
    state.serviceDialog.title = `${row.name} => ${authCert.username}@${row.ip}`;
};

/**
 * 显示机器状态统计信息
 */
const showMachineStats = (machine: any) => {
    state.machineStatsDialog.machineId = machine.id;
    state.machineStatsDialog.title = `${t('machine.machineState')}: ${machine.name} => ${machine.ip}`;
    state.machineStatsDialog.visible = true;
};

const showFileManage = (selectionData: any) => {
    state.fileDialog.machine = selectionData;
    state.fileDialog.visible = true;
};

/**
 * 处理文件配置选择事件
 */
const onFileConfigSelect = (fileConfig: { fileId: number; path: string; name: string; type: number }) => {
    const machine = state.fileDialog.machine;
    if (!machine) return;

    // 获取当前机器信息
    const machineId = machine.id;
    const authCertName = machine.selectAuthCert.name;

    // 生成文件操作 tab 的 key
    const fileTabKey = `file_${machineId}_${authCertName}_${fileConfig.fileId}`;

    // 检查是否已经存在该文件操作 tab
    if (state.tabs.has(fileTabKey)) {
        // 如果已存在，直接切换到该 tab
        state.activeTermName = fileTabKey;
        return;
    }

    // 使用国际化前缀拼接 tab 标签
    const labelName = `${t('machine.fileTabPrefix')}${machine.selectAuthCert.username}@${machine.name}/${fileConfig.name}`;

    let tab = {
        key: fileTabKey,
        label: labelName.length > 25 ? labelName.slice(0, 18) + '...' + labelName.slice(-7) : labelName,
        type: 'file',
        machineId: machineId,
        authCertName: authCertName,
        protocol: machine.protocol,
        fileId: fileConfig.fileId,
        path: fileConfig.path,
        params: machine,
    };

    state.tabs.set(fileTabKey, tab);
    state.activeTermName = fileTabKey;
};

const showInfo = (info: any) => {
    state.infoDialog.data = info;
    state.infoDialog.visible = true;
};

const showProcess = (row: any) => {
    state.processDialog.machineId = row.id;
    state.processDialog.visible = true;
};

const showRec = (row: any) => {
    state.machineRecDialog.title = `${row.name}[${row.ip}]-${t('machine.terminalPlayback')}`;
    state.machineRecDialog.machineId = row.id;
    state.machineRecDialog.visible = true;
};

const onRemoveTab = (targetName: string) => {
    let activeTermName = state.activeTermName;
    const tabNames = [...state.tabs.keys()];
    for (let i = 0; i < tabNames.length; i++) {
        const tabName = tabNames[i];
        if (tabName !== targetName) {
            continue;
        }

        const tab = state.tabs.get(targetName);

        // 只有终端类型才需要关闭连接
        if (tab?.type === 'terminal') {
            terminalRefs[targetName]?.close();
        }

        state.tabs.delete(targetName);

        if (activeTermName != targetName) {
            break;
        }

        // 如果删除的 tab 是当前激活的 tab，则切换到前一个或后一个 tab
        const nextTab = tabNames[i + 1] || tabNames[i - 1];
        if (nextTab) {
            activeTermName = nextTab;
        } else {
            activeTermName = '';
        }

        state.activeTermName = activeTermName;
        break;
    }
};

const terminalStatusChange = (key: string, status: TerminalStatus) => {
    state.tabs.get(key).status = status;
};

const terminalRefs: any = {};
const setTerminalRef = (el: any, key: any) => {
    if (key) {
        terminalRefs[key] = el;
    }
};

const fitTerminal = () => {
    setTimeout(() => {
        let info = state.tabs.get(state.activeTermName);
        // 只有终端类型才需要调整大小
        if (info && info.type === 'terminal') {
            terminalRefs[info.key]?.fitTerminal && terminalRefs[info.key]?.fitTerminal();
        }
    });
};

const handleReconnect = (tab: any, force = false) => {
    // 只有终端类型才需要重连
    if (tab?.type === 'terminal') {
        terminalRefs[tab.key]?.init();
    }
};

defineExpose({
    openTerminal,
    onResize: fitTerminal,
    showInfo,
    showProcess,
    showRec,
    showMachineStats,
    showFileManage,
    serviceManager,
});
</script>

<style lang="scss">
.machine-terminal-tabs {
    --el-tabs-header-height: 30px;

    .el-tabs {
        --el-tabs-header-height: 30px;
    }

    .machine-terminal-tab-label {
        font-size: 12px;
    }

    .el-tabs__header {
        margin-bottom: 5px;
    }

    .el-tabs__item {
        padding: 0 8px !important;
    }
}
</style>
