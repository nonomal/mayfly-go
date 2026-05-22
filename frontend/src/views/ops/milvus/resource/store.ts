import { defineStore } from 'pinia';
import { milvusApi } from '@/views/ops/milvus/api';
import { setCurrentAcName } from './authCert';

/**
 * 缓存milvus一些参数
 */
export const useMilvusStore = defineStore('milvusStore', {
    state: (): MilvusState => ({
        dbs: [],
        selectedDb: 'default',
        collections: [],
        selectedCollection: '',
        authCertName: '',
    }),
    actions: {
        setDbs(dbs: any[]) {
            this.dbs = dbs;
        },
        async refreshDbs(milvusId: number) {
            const res = await milvusApi.listDatabases(milvusId);
            // res 通过dbid排序
            res.sort((a: any, b: any) => {
                return a.create_time.localeCompare(b.create_time);
            });
            this.dbs = res;
            if (res.length > 0) {
                this.selectedDb = res[0].name;
                milvusApi.useDatabase(res[0].id, res[0].name);
            }
        },
        setSelectedDb(db: string) {
            this.selectedDb = db;
        },
        setSelectedCollection(coll: string) {
            console.log('[MilvusStore] 切换 collection:', coll);
            this.selectedCollection = coll;
        },
        setCollections(collections: string[]) {
            collections.sort();
            this.collections = collections;
            // 默认选中第一个 collection
            if (!this.selectedCollection && this.collections.length > 0) {
                this.setSelectedCollection(this.collections[0]);
            }
        },
        clear() {
            console.log('[MilvusStore] 清空状态');
            this.collections = [];
            this.selectedCollection = '';
            this.selectedDb = 'default';
            this.dbs = [];
        },
        setAuthCertName(name: string) {
            this.authCertName = name;
            setCurrentAcName(name);
        },
    },
});