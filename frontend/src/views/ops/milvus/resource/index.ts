import { defineAsyncComponent } from 'vue';
import { NodeType, TagTreeNode, ResourceComponentConfig, ResourceConfig } from '../../component/tag';
import { ResourceTypeEnum, TagResourceTypeEnum } from '@/common/commonEnum';
import { sleep } from '@/common/utils/loading';
import { milvusApi, perms } from '@/views/ops/milvus/api';

export const MilvusIcon = {
    name: ResourceTypeEnum.Milvus.extra.icon,
    color: ResourceTypeEnum.Milvus.extra.iconColor,
};

const MilvusList = defineAsyncComponent(() => import('../MilvusList.vue'));
const MilvusOp = defineAsyncComponent(() => import('./MilvusOp.vue'));

const NodeMilvus = defineAsyncComponent(() => import('./NodeMilvus.vue'));

export const MilvusOpComp: ResourceComponentConfig = {
    name: 'tag.milvusOp',
    component: MilvusOp,
    icon: MilvusIcon,
};

const NodeTypeMilvus = new NodeType(TagResourceTypeEnum.Milvus.value).withNodeClickFunc(async (node: TagTreeNode) => {
    (await node.ctx?.addResourceComponent(MilvusOpComp))?.initMilvus(node.params);
});

// tagpath 节点类型
const NodeTypeMilvusTag = new NodeType(TagTreeNode.TagPath).withLoadNodesFunc(async (parentNode: TagTreeNode) => {
    const tagPath = parentNode.params.tagPath;
    const res = await milvusApi.list.request({ tagPath });
    if (!res.total) {
        return [];
    }
    const milvusInfos = res.list;
    await sleep(100);
    return milvusInfos.map((x: any) => {
        return TagTreeNode.new(parentNode, `${x.code}`, x.name, NodeTypeMilvus).withIsLeaf(true).withParams(x).withNodeComponent(NodeMilvus);
    });
});

export default {
    order: 7,
    resourceType: TagResourceTypeEnum.Milvus.value,
    rootNodeType: NodeTypeMilvusTag,
    manager: {
        componentConf: {
            component: MilvusList,
            icon: MilvusIcon,
            name: 'milvus',
        },
        countKey: 'milvus',
        permCode: perms.base,
    },
} as ResourceConfig;
