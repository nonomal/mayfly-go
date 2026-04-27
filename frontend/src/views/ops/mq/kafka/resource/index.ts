import { defineAsyncComponent } from 'vue';
import { NodeType, TagTreeNode, ResourceComponentConfig, ResourceConfig } from '../../../component/tag';
import { ResourceTypeEnum, TagResourceTypeEnum } from '@/common/commonEnum';
import { sleep } from '@/common/utils/loading';
import { mqApi } from '@/views/ops/mq/api';

export const KafkaIcon = {
    name: ResourceTypeEnum.MqKafka.extra.icon,
    color: ResourceTypeEnum.MqKafka.extra.iconColor,
};

const KafkaList = defineAsyncComponent(() => import('../KafkaList.vue'));
const KafkaOp = defineAsyncComponent(() => import('./KafkaOp.vue'));

const NodeKafka = defineAsyncComponent(() => import('./NodeKafka.vue'));

export const KafkaOpComp: ResourceComponentConfig = {
    name: 'tag.mq.kafkaOp',
    component: KafkaOp,
    icon: KafkaIcon,
};

const NodeTypeKafka = new NodeType(TagResourceTypeEnum.MqKafka.value).withNodeClickFunc(async (node: TagTreeNode) => {
    (await node.ctx?.addResourceComponent(KafkaOpComp)).initKafka(node.params);
});

// tagpath 节点类型
const NodeTypeKafkaTag = new NodeType(TagTreeNode.TagPath).withLoadNodesFunc(async (parentNode: TagTreeNode) => {
    const tagPath = parentNode.params.tagPath;
    const res = await mqApi.kafkaList.request({ tagPath });
    if (!res.total) {
        return [];
    }
    const kafkaInfos = res.list;
    await sleep(100);
    return kafkaInfos.map((x: any) => {
        return TagTreeNode.new(parentNode, `${x.code}`, x.name, NodeTypeKafka).withIsLeaf(true).withParams(x).withNodeComponent(NodeKafka);
    });
});

export default {
    order: 6.1,
    resourceType: TagResourceTypeEnum.MqKafka.value,
    rootNodeType: NodeTypeKafkaTag,
    manager: {
        componentConf: {
            component: KafkaList,
            icon: KafkaIcon,
            name: 'kafka',
        },
        countKey: 'kafka',
        permCode: 'mq:kafka:base',
    },
} as ResourceConfig;
