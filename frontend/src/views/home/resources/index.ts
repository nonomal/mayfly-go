import Database from './Database.vue';
import Docker from './Docker.vue';
import ES from './ES.vue';
import Kafka from './Kafka.vue';
import Machine from './Machine.vue';
import Milvus from './Milvus.vue';
import Mongo from './Mongo.vue';
import Redis from './Redis.vue';

export const resourceComponents = [Machine, Database, Redis, Mongo, ES, Milvus, Docker, Kafka];
