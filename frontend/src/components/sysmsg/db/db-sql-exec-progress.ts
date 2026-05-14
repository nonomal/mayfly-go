import ProgressNotify from './DbSqlExecProgress.vue';
import { ElNotification } from 'element-plus';
import { h, reactive } from 'vue';
import syssocket from '@/common/syssocket';

const sqlExecNotifyMap: Map<string, any> = new Map();

// 构建 props（私有函数，不导出）
const buildProgressProps = (): any => {
    return {
        progress: {
            title: {
                type: String,
            },
            executedStatements: {
                type: Number,
            },
        },
    };
};

export async function registerDbSqlExecProgress() {
    await syssocket.registerMsgHandler('sqlScriptRunProgress', function (message: any) {
        const content = message.params;
        const id = content.id;
        let progress = sqlExecNotifyMap.get(id);
        if (content.terminated) {
            if (progress != undefined) {
                progress.notification?.close();
                sqlExecNotifyMap.delete(id);
                progress = undefined;
            }
            return;
        }

        if (progress == undefined) {
            progress = {
                props: reactive(buildProgressProps()),
                notification: undefined,
            };
        }

        progress.props.progress.title = content.title;
        progress.props.progress.executedStatements = content.executedStatements;
        if (!sqlExecNotifyMap.has(id)) {
            progress.notification = ElNotification({
                duration: 0,
                title: message.title,
                message: h(ProgressNotify, progress.props),
                type: 'info',
                showClose: false,
            });
            sqlExecNotifyMap.set(id, progress);
        }
    });
}
