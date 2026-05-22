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
const NodeMilvusAc = defineAsyncComponent(() => import('./NodeMilvusAc.vue'));

export const MilvusOpComp: ResourceComponentConfig = {
    name: 'tag.milvusOp',
    component: MilvusOp,
    icon: MilvusIcon,
};

// milvus 授权凭证节点类型
const NodeTypeMilvusAc = new NodeType(TagResourceTypeEnum.Milvus.value * 10 + 1).withNodeClickFunc(async (node: TagTreeNode) => {

    (await node.ctx?.addResourceComponent(MilvusOpComp))?.initMilvus(node.params);

    console.log(node.params);
});

const NodeTypeMilvus = new NodeType(TagResourceTypeEnum.Milvus.value).withLoadNodesFunc((node: TagTreeNode) => {
    const milvus = node.params;
    const authCerts = milvus.authCerts || [];
    return authCerts.map((x: any) =>
        TagTreeNode.new(node, x.name, x.username, NodeTypeMilvusAc)
            .withNodeComponent(NodeMilvusAc)
            .withParams({ ...milvus, selectAuthCert: x })
            .withIsLeaf(true)
            .withIcon({ name: 'Ticket', color: '#409eff' })
    );
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
        return TagTreeNode.new(parentNode, `${x.code}`, x.name, NodeTypeMilvus).withParams(x).withNodeComponent(NodeMilvus);
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
