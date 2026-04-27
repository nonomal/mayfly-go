<template>
    <div class="create-collection-drawer">
        <el-drawer
            :title="drawerTitle"
            v-model="visible"
            :before-close="handleClose"
            :destroy-on-close="false"
            :close-on-click-modal="false"
            :append-to-body="false"
            size="80%"
        >
            <el-form ref="formRef" :model="form" :rules="rules" label-width="auto">
                <!-- 基本信息 -->
                <el-divider content-position="left">{{ $t('common.basic') }}</el-divider>
                <el-row :gutter="20">
                    <el-col :span="10">
                        <el-form-item :label="$t('common.name')" prop="name">
                            <el-input v-model="form.name" :placeholder="$t('milvus.collectionNamePlaceholder')"></el-input>
                        </el-form-item>
                    </el-col>
                    <el-col :span="6">
                        <el-form-item :label="$t('milvus.shardsNum')" prop="shardsNum">
                            <el-input-number v-model="form.shardsNum" :min="1" :max="64" style="width: 100%" />
                        </el-form-item>
                    </el-col>
                    <el-col :span="8">
                        <el-form-item :label="$t('milvus.consistencyLevel')" prop="consistency_level">
                            <el-select v-model="form.consistency_level">
                                <el-option label="Bounded" value="Bounded" />
                                <el-option label="Strong" value="Strong" />
                                <el-option label="Session" value="Session" />
                                <el-option label="Eventually" value="Eventually" />
                                <!--                      <el-option label="Customized" value="Customized" />-->
                            </el-select>
                        </el-form-item>
                    </el-col>
                </el-row>
                <el-form-item :label="$t('milvus.description')" prop="description">
                    <el-input v-model="form.description" type="textarea" :rows="2" :placeholder="$t('milvus.descriptionPlaceholder')"></el-input>
                </el-form-item>

                <!-- Schema 配置 -->
                <el-divider content-position="left">Schema</el-divider>

                <div class="schema-container">
                    <!-- 左侧：字段列表 -->
                    <div class="schema-left">
                        <div class="schema-header">
                            <span class="schema-title">{{ $t('milvus.fields') }}</span>
                            <el-space>
                                <!-- 动态字段开关 -->
                                <el-tooltip :content="dynamicFieldEnabled ? $t('milvus.disableDynamicField') : $t('milvus.enableDynamicField')" placement="top">
                                    <el-switch
                                        v-model="dynamicFieldEnabled"
                                        :disabled="isEditMode"
                                        @change="handleDynamicFieldToggle"
                                        active-text=""
                                        inactive-text=""
                                    >
                                        <template #active-action>
                                            <el-icon><check /></el-icon>
                                        </template>
                                        <template #inactive-action>
                                            <el-icon><close /></el-icon>
                                        </template>
                                    </el-switch>
                                </el-tooltip>
                                <el-popover
                                    v-model:visible="fieldTypePopoverVisible"
                                    placement="right-start"
                                    :width="400"
                                    trigger="click"
                                    popper-class="field-type-popover"
                                >
                                    <template #reference>
                                        <el-button type="primary" size="small">
                                            {{ $t('common.add') }}{{ $t('milvus.field') }}
                                            <el-icon class="el-icon--right"><arrow-down /></el-icon>
                                        </el-button>
                                    </template>

                                    <div class="field-type-selector">
                                        <!-- 左侧：分类列表 -->
                                        <div class="category-list">
                                            <div class="category-list-header">分类</div>
                                            <div class="category-list-body">
                                                <div
                                                    v-for="(types, category) in fieldTypeGroups"
                                                    :key="category"
                                                    class="category-item"
                                                    :class="{ active: selectedCategory === category }"
                                                    @mouseenter="selectedCategory = category"
                                                >
                                                    {{ category }}
                                                </div>
                                            </div>
                                        </div>
                                        <!-- 右侧：类型列表 -->
                                        <div class="type-list">
                                            <div class="type-list-header">类型</div>
                                            <div class="type-list-body">
                                                <div
                                                    v-for="type in fieldTypeGroups[selectedCategory]"
                                                    :key="type.value"
                                                    class="type-item"
                                                    @click="handleSelectFieldType(type.value)"
                                                >
                                                    {{ type.label }}
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </el-popover>
                            </el-space>
                        </div>

                        <div class="field-list">
                            <!-- 普通字段列表 -->
                            <div
                                v-for="(field, index) in normalFields"
                                :key="'normal-' + index"
                                class="field-item"
                                :class="{ active: selectedFieldIndex === index, readonly: isEditMode && field.readonly }"
                                @click="selectField(index)"
                            >
                                <div class="field-icon">
                                    <el-icon v-if="field.isPrimaryKey"><key /></el-icon>
                                    <el-icon v-else-if="isVectorType(field.dataType)"><trend-charts /></el-icon>
                                    <el-icon v-else-if="isGeometryType(field.dataType)"><location /></el-icon>
                                    <el-icon v-else-if="isClockType(field.dataType)"><clock /></el-icon>
                                    <el-icon v-else-if="isArrayType(field.dataType)"><memo /></el-icon>
                                    <el-icon v-else><document /></el-icon>
                                </div>
                                <div class="field-info">
                                    <div class="field-name">
                                        <span class="field-name-text">{{ field.name }}</span>
                                        <el-tag v-if="field.isPrimaryKey" size="small" type="primary">PK</el-tag>
                                        <el-tag v-else-if="field.isPartitionKey" size="small" type="success">Partition Key</el-tag>
                                        <el-tag v-if="field.readonly" size="small" type="info">{{ $t('milvus.readonly') }}</el-tag>
                                    </div>
                                    <div class="field-type">{{ formatFieldType(field) }}</div>
                                </div>
                                <div class="field-index-tag">
                                    <el-tag v-if="field.indexType" size="small" type="success">{{ field.indexType }}</el-tag>
                                    <el-tag v-else size="small" type="info">{{ $t('milvus.noIndex') }}</el-tag>
                                </div>
                                <div class="field-actions">
                                    <el-space>
                                        <el-button
                                            text
                                            size="small"
                                            type="primary"
                                            @click.stop="handleCopyField(index)"
                                            :disabled="(field.isPrimaryKey && field.autoID) || (isEditMode && field.readonly)"
                                        >
                                            <el-icon><copy-document /></el-icon>
                                        </el-button>
                                        <el-button
                                            text
                                            size="small"
                                            type="danger"
                                            @click.stop="handleDeleteField(index)"
                                            :disabled="(field.isPrimaryKey && field.autoID) || (isEditMode && field.readonly)"
                                        >
                                            <el-icon><delete /></el-icon>
                                        </el-button>
                                    </el-space>
                                </div>
                            </div>

                            <!-- 动态字段 -->
                            <div
                                v-if="dynamicFieldEnabled && dynamicField"
                                class="field-item dynamic-field"
                                :class="{ active: selectedFieldIndex === -2 }"
                                @click="selectDynamicField"
                            >
                                <div class="field-icon">
                                    <el-icon><circle-plus /></el-icon>
                                </div>
                                <div class="field-info">
                                    <div class="field-name">
                                        <span class="field-name-text">$meta</span>
                                        <el-tag size="small" type="warning">{{ $t('milvus.dynamicField') }}</el-tag>
                                    </div>
                                    <div class="field-type">JSON</div>
                                </div>
                                <div class="field-index-tag">
                                    <el-tag v-if="dynamicField.indexes && dynamicField.indexes.length > 0" size="small" type="success">
                                        {{ dynamicField.indexes.length }} {{ $t('milvus.indexes') }}
                                    </el-tag>
                                    <el-tag v-else size="small" type="info">{{ $t('milvus.noIndex') }}</el-tag>
                                </div>
                                <div class="field-actions">
                                    <el-space>
                                        <el-button text size="small" type="danger" @click.stop="handleDeleteDynamicField" :disabled="isEditMode">
                                            <el-icon><delete /></el-icon>
                                        </el-button>
                                    </el-space>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- 右侧：字段属性配置 -->
                    <div class="schema-right">
                        <el-tabs v-model="activeTab" v-if="selectedField || (dynamicFieldEnabled && dynamicField)">
                            <el-tab-pane :label="$t('milvus.properties')" name="properties">
                                <!-- 普通字段属性 -->
                                <el-form :model="selectedField" label-width="auto" v-if="selectedField && !isDynamicFieldSelected">
                                    <el-form-item :label="$t('milvus.field')" required>
                                        <el-input
                                            v-model="selectedField.name"
                                            :placeholder="$t('milvus.fieldNamePlaceholder')"
                                            :disabled="isEditMode && selectedField.readonly"
                                        ></el-input>
                                    </el-form-item>

                                    <el-form-item :label="$t('milvus.dataType')">
                                        <el-select v-model="selectedField.dataType" filterable :disabled="isEditMode && selectedField.readonly">
                                            <el-option-group v-for="(types, category) in fieldTypeGroups" :key="category" :label="category">
                                                <el-option v-for="type in types" :key="type.value" :label="type.label" :value="type.value"></el-option>
                                            </el-option-group>
                                        </el-select>
                                    </el-form-item>

                                    <!-- 向量类型特有属性 -->
                                    <template v-if="isVectorType(selectedField.dataType)">
                                        <el-form-item :label="$t('milvus.dim')" prop="dim">
                                            <el-input-number v-model="selectedField.dim" :min="1" style="width: 100%" />
                                        </el-form-item>
                                    </template>

                                    <!-- VarChar 类型特有属性 -->
                                    <template v-if="isVarCharType(selectedField.dataType)">
                                        <el-form-item :label="$t('milvus.maxLength')" prop="maxLength">
                                            <el-input-number v-model="selectedField.maxLength" :min="1" :max="65535" style="width: 100%" />
                                        </el-form-item>
                                    </template>

                                    <!-- 数组类型特有属性 -->
                                    <template v-if="isArrayType(selectedField.dataType)">
                                        <el-form-item :label="$t('milvus.elementType')" prop="elementType">
                                            <el-select v-model="selectedField.elementType" style="width: 100%" filterable>
                                                <el-option label="Bool" value="Bool" />
                                                <el-option label="Int8" value="Int8" />
                                                <el-option label="Int16" value="Int16" />
                                                <el-option label="Int32" value="Int32" />
                                                <el-option label="Int64" value="Int64" />
                                                <el-option label="Float" value="Float" />
                                                <el-option label="Double" value="Double" />
                                                <el-option label="VarChar" value="VarChar" />
                                                <el-option label="Struct" value="Struct" />
                                            </el-select>
                                        </el-form-item>
                                        <el-form-item :label="$t('milvus.maxCapacity')" prop="maxCapacity">
                                            <el-input-number v-model="selectedField.maxCapacity" :min="1" style="width: 100%" />
                                        </el-form-item>
                                    </template>

                                    <el-divider />

                                    <el-form-item :label="$t('milvus.primaryKey')">
                                        <el-checkbox
                                            v-model="selectedField.isPrimaryKey"
                                            :disabled="
                                                (isEditMode && selectedField.readonly) ||
                                                    !canBePrimaryKey(selectedField.dataType) ||
                                                    (!!getPrimaryKeyField() && getPrimaryKeyField() !== selectedField)
                                            "
                                        />
                                        <el-tooltip
                                            v-if="!canBePrimaryKey(selectedField.dataType)"
                                            :content="$t('milvus.primaryKeyOnlyInt64OrVarChar')"
                                            placement="top"
                                        >
                                            <el-icon style="margin-left: 4px; color: var(--el-color-warning)"><warning-filled /></el-icon>
                                        </el-tooltip>
                                    </el-form-item>

                                    <el-form-item v-if="selectedField.isPrimaryKey && isInt64Type(selectedField.dataType)" :label="$t('milvus.autoID')">
                                        <el-switch v-model="selectedField.autoID" :disabled="isEditMode && selectedField.readonly" />
                                    </el-form-item>

                                    <el-form-item :label="$t('milvus.nullable')">
                                        <el-checkbox v-model="selectedField.nullable" />
                                    </el-form-item>

                                    <el-form-item :label="$t('milvus.partitionKey')">
                                        <el-switch v-model="selectedField.isPartitionKey" />
                                    </el-form-item>

                                    <el-form-item :label="$t('milvus.clusteringKey')">
                                        <el-switch v-model="selectedField.isClusteringKey" />
                                    </el-form-item>

                                    <el-form-item :label="$t('milvus.mmap')">
                                        <el-switch v-model="selectedField.mmap" /> <el-icon> <question-filled /></el-icon>
                                    </el-form-item>

                                    <el-divider />

                                    <el-form-item :label="$t('milvus.defaultValue')">
                                        <el-input v-model="selectedField.defaultValue" :placeholder="$t('milvus.defaultValuePlaceholder')"></el-input>
                                    </el-form-item>

                                    <el-form-item :label="$t('milvus.description')">
                                        <el-input
                                            v-model="selectedField.description"
                                            type="textarea"
                                            :rows="3"
                                            :placeholder="$t('milvus.descriptionPlaceholder')"
                                        ></el-input>
                                    </el-form-item>
                                </el-form>

                                <!-- 动态字段属性 -->
                                <el-form :model="dynamicField" label-width="auto" v-if="isDynamicFieldSelected">
                                    <el-form-item :label="$t('milvus.dynamicField')">
                                        <el-tag type="warning">$meta (JSON)</el-tag>
                                    </el-form-item>
                                    <el-alert :title="$t('milvus.dynamicFieldHint')" type="info" :closable="false" show-icon />
                                </el-form>
                            </el-tab-pane>

                            <el-tab-pane :label="$t('milvus.index')" name="index">
                                <!-- 动态字段索引配置 -->
                                <div v-if="isDynamicFieldSelected" class="index-config">
                                    <div>
                                        <div>
                                            <el-space>
                                                <span class="index-title">{{ $t('milvus.dynamicFieldIndexes') }}</span>
                                                <el-button type="primary" size="small" @click="handleAddDynamicIndex">
                                                    <el-icon><plus /></el-icon>
                                                    {{ $t('milvus.addIndex') }}
                                                </el-button>
                                            </el-space>
                                        </div>

                                        <div v-if="dynamicField.indexes.length === 0" class="index-empty">
                                            <el-empty :description="$t('milvus.noIndexCreated')" :image-size="60" />
                                        </div>

                                        <div v-else>
                                            <el-space direction="vertical" alignment="stretch">
                                                <!-- 索引列表 -->
                                                <div class="index-list-tabs">
                                                    <div
                                                        v-for="(idx, idxIdx) in dynamicField.indexes"
                                                        :key="idxIdx"
                                                        class="index-list-tab"
                                                        :class="{ active: dynamicField.selectedIdx === idxIdx }"
                                                        @click="dynamicField.selectedIdx = idxIdx"
                                                    >
                                                        <span>{{ idx.indexName || `${$t('common.index')} ${Number(idxIdx) + 1}` }}</span>
                                                        <el-popconfirm :title="$t('milvus.confirmDeleteIndex')" @confirm="handleDeleteDynamicIndex(Number(idxIdx))">
                                                            <template #reference>
                                                                <el-icon class="delete-icon"><close /></el-icon>
                                                            </template>
                                                        </el-popconfirm>
                                                    </div>
                                                </div>

                                                <!-- 当前选中索引的配置 -->
                                                <el-form
                                                    v-if="dynamicField.indexes[dynamicField.selectedIdx]"
                                                    :model="dynamicField.indexes[dynamicField.selectedIdx]"
                                                    label-width="auto"
                                                >
                                                    <el-form-item :label="$t('milvus.indexName')">
                                                        <el-input
                                                            v-model="dynamicField.indexes[dynamicField.selectedIdx].indexName"
                                                            :placeholder="$t('milvus.indexNamePlaceholder')"
                                                        />
                                                    </el-form-item>

                                                    <el-form-item :label="$t('milvus.indexType')">
                                                        <el-select
                                                            v-model="dynamicField.indexes[dynamicField.selectedIdx].indexType"
                                                            style="width: 100%"
                                                            filterable
                                                        >
                                                            <el-option label="AUTOINDEX" value="AUTOINDEX" />
                                                            <el-option label="INVERTED" value="INVERTED" />
                                                            <el-option label="Trie" value="Trie" />
                                                            <el-option label="STL_SORT" value="STL_SORT" />
                                                            <el-option label="BITMAP" value="BITMAP" />
                                                        </el-select>
                                                    </el-form-item>

                                                    <el-form-item :label="$t('milvus.jsonPath')" required>
                                                        <el-input
                                                            v-model="dynamicField.indexes[dynamicField.selectedIdx].json_path"
                                                            :placeholder="$t('milvus.jsonPathPlaceholder')"
                                                        />
                                                        <div class="form-tip">{{ $t('milvus.jsonPathTip') }}</div>
                                                    </el-form-item>

                                                    <el-form-item :label="$t('milvus.jsonCastType')" required>
                                                        <el-select v-model="dynamicField.indexes[dynamicField.selectedIdx].json_cast_type" style="width: 100%">
                                                            <el-option label="varchar" value="varchar" />
                                                            <el-option label="double" value="double" />
                                                            <el-option label="bool" value="bool" />
                                                            <el-option label="array_varchar" value="array_varchar" />
                                                            <el-option label="array_double" value="array_double" />
                                                            <el-option label="array_bool" value="array_bool" />
                                                        </el-select>
                                                        <div class="form-tip">{{ $t('milvus.jsonCastTypeTip') }}</div>
                                                    </el-form-item>

                                                    <el-form-item :label="$t('milvus.jsonCastFunction')">
                                                        <el-input
                                                            v-model="dynamicField.indexes[dynamicField.selectedIdx].json_cast_function"
                                                            :placeholder="$t('milvus.jsonCastFunctionPlaceholder')"
                                                        />
                                                        <div class="form-tip">{{ $t('milvus.jsonCastFunctionTip') }}</div>
                                                    </el-form-item>

                                                    <el-form-item :label="$t('milvus.fieldName')" required>
                                                        <el-input
                                                            v-model="dynamicField.indexes[dynamicField.selectedIdx].field_name"
                                                            :placeholder="$t('milvus.fieldNamePlaceholder')"
                                                        />
                                                        <div class="form-tip">{{ $t('milvus.dynamicFieldNameTip') }}</div>
                                                    </el-form-item>
                                                </el-form>
                                            </el-space>
                                        </div>
                                    </div>
                                </div>

                                <!-- 普通字段索引配置 -->
                                <div v-else-if="selectedField" class="index-config">
                                    <el-alert
                                        v-if="!supportsIndex(selectedField.dataType)"
                                        :title="$t('milvus.noIndexSupport')"
                                        type="info"
                                        :closable="false"
                                        show-icon
                                    />
                                    <div v-else>
                                        <!-- 未创建索引时显示创建按钮 -->
                                        <el-empty v-if="!selectedField.indexType" :description="$t('milvus.noIndexCreated')" :image-size="80">
                                            <el-button type="primary" @click="handleAddIndex(selectedField)">
                                                {{ $t('milvus.createIndex') }}
                                            </el-button>
                                        </el-empty>

                                        <!-- 已创建索引时显示索引配置 -->
                                        <div v-else class="index-detail">
                                            <div class="index-header">
                                                <span class="index-title">{{ $t('milvus.currentIndex') }}</span>
                                                <el-popconfirm
                                                    v-if="!isEditMode"
                                                    :title="$t('milvus.confirmDeleteIndex')"
                                                    @confirm="handleRemoveIndex(selectedField)"
                                                >
                                                    <template #reference>
                                                        <el-button type="danger" size="small" text>
                                                            <el-icon><delete /></el-icon>
                                                            {{ $t('milvus.deleteIndex') }}
                                                        </el-button>
                                                    </template>
                                                </el-popconfirm>
                                            </div>

                                            <el-form :model="selectedField" label-width="auto" class="index-form">
                                                <el-form-item :label="$t('milvus.indexType')">
                                                    <el-select v-model="selectedField.indexType" style="width: 100%" filterable :disabled="isEditMode">
                                                        <el-option-group
                                                            v-for="(indexes, group) in getIndexesByGroup(selectedField.dataType)"
                                                            :key="group"
                                                            :label="group"
                                                        >
                                                            <el-option v-for="idx in indexes" :key="idx.value" :label="idx.label" :value="idx.value" />
                                                        </el-option-group>
                                                    </el-select>
                                                </el-form-item>

                                                <!-- 向量类型显示度量类型 -->
                                                <el-form-item v-if="getMetricOptions(selectedField.dataType).length > 0" :label="$t('milvus.metricType')">
                                                    <el-select v-model="selectedField.metricType" style="width: 100%" :disabled="isEditMode">
                                                        <el-option
                                                            v-for="metric in getMetricOptions(selectedField.dataType)"
                                                            :key="metric"
                                                            :label="metric"
                                                            :value="metric"
                                                        />
                                                    </el-select>
                                                </el-form-item>

                                                <!-- HNSW 参数 -->
                                                <template v-if="selectedField.indexType?.startsWith('HNSW')">
                                                    <el-form-item label="M">
                                                        <el-input-number
                                                            v-model="selectedField.indexParams.M"
                                                            :min="1"
                                                            :max="200"
                                                            style="width: 100%"
                                                            :disabled="isEditMode"
                                                        />
                                                    </el-form-item>
                                                    <el-form-item label="efConstruction">
                                                        <el-input-number
                                                            v-model="selectedField.indexParams.efConstruction"
                                                            :min="1"
                                                            :max="65535"
                                                            style="width: 100%"
                                                            :disabled="isEditMode"
                                                        />
                                                    </el-form-item>
                                                    <el-form-item
                                                        v-if="selectedField.indexType?.includes('PQ') || selectedField.indexType?.includes('PRQ')"
                                                        label="m"
                                                    >
                                                        <el-input-number
                                                            v-model="selectedField.indexParams.m"
                                                            :min="1"
                                                            style="width: 100%"
                                                            :disabled="isEditMode"
                                                        />
                                                    </el-form-item>
                                                    <el-form-item
                                                        v-if="selectedField.indexType?.includes('PQ') || selectedField.indexType?.includes('PRQ')"
                                                        label="nbits"
                                                    >
                                                        <el-input-number
                                                            v-model="selectedField.indexParams.nbits"
                                                            :min="1"
                                                            :max="16"
                                                            style="width: 100%"
                                                            :disabled="isEditMode"
                                                        />
                                                    </el-form-item>
                                                </template>

                                                <!-- IVF 参数 -->
                                                <template v-if="selectedField.indexType?.startsWith('IVF') || selectedField.indexType === 'SCANN'">
                                                    <el-form-item label="nlist">
                                                        <el-input
                                                            v-model="selectedField.indexParams.nlist"
                                                            :min="1"
                                                            :max="65536"
                                                            style="width: 100%"
                                                            :disabled="isEditMode"
                                                        />
                                                    </el-form-item>
                                                    <el-form-item v-if="selectedField.indexType?.includes('PQ')" label="m">
                                                        <el-input v-model="selectedField.indexParams.m" :min="1" style="width: 100%" :disabled="isEditMode" />
                                                    </el-form-item>
                                                    <el-form-item v-if="selectedField.indexType?.includes('PQ')" label="nbits">
                                                        <el-input
                                                            v-model="selectedField.indexParams.nbits"
                                                            :min="1"
                                                            :max="16"
                                                            style="width: 100%"
                                                            :disabled="isEditMode"
                                                        />
                                                    </el-form-item>
                                                </template>

                                                <!-- SPARSE_INVERTED_INDEX 参数 -->
                                                <template v-if="selectedField.indexType === 'SPARSE_INVERTED_INDEX'">
                                                    <el-form-item label="drop_ratio_build">
                                                        <el-input-number
                                                            v-model="selectedField.indexParams.drop_ratio_build"
                                                            :min="0"
                                                            :max="1"
                                                            :step="0.1"
                                                            style="width: 100%"
                                                            :disabled="isEditMode"
                                                        />
                                                    </el-form-item>
                                                </template>
                                            </el-form>
                                        </div>
                                    </div>
                                </div>
                            </el-tab-pane>
                        </el-tabs>

                        <el-empty v-else :description="$t('milvus.selectFieldToConfig')" />
                    </div>
                </div>
            </el-form>

            <template #footer>
                <div class="drawer-footer">
                    <el-button @click="handleClose">{{ $t('common.cancel') }}</el-button>
                    <el-button type="primary" @click="handleSubmit" :loading="loading">{{ $t('common.confirm') }}</el-button>
                    <el-button type="primary" @click="handlePreview" :loading="loading">{{ 'Json' + $t('common.preview') }}</el-button>
                </div>
            </template>
        </el-drawer>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { ElMessage, FormInstance } from 'element-plus';
import {
    ArrowDown,
    Key,
    TrendCharts,
    Document,
    Delete,
    Location,
    Clock,
    Memo,
    CopyDocument,
    QuestionFilled,
    WarningFilled,
    Check,
    Close,
    Plus,
    CirclePlus,
} from '@element-plus/icons-vue';
import DrawerHeader from '@/components/drawer-header/DrawerHeader.vue';
import { milvusApi } from '../api';
import { Rules } from '@/common/rule';
import { useI18n } from 'vue-i18n';
import MonacoEditorBox from '@/components/monaco/MonacoEditorBox';

const { t } = useI18n();

const props = defineProps<{
    milvusId: number;
    mode?: 'create' | 'edit' | 'copy';
    editData?: any;
}>();

const emits = defineEmits(['success']);

const visible = defineModel<boolean>('visible', { default: false });

const formRef = ref<FormInstance>();
const loading = ref(false);
const activeTab = ref('properties');
const selectedFieldIndex = ref<number>(-1);
const fieldTypePopoverVisible = ref(false);
const selectedCategory = ref<any>('向量');
const dynamicFieldEnabled = ref(false);

// 动态字段
const dynamicField = ref<any>({
    indexes: [] as any[], // 多个索引
    selectedIdx: 0, // 当前选中的索引
});

// 计算属性：普通字段列表
const normalFields = computed(() => {
    return form.value.fields || [];
});

// 是否选中动态字段
const isDynamicFieldSelected = computed(() => selectedFieldIndex.value === -2);

// 计算抽屉标题
const drawerTitle = computed(() => {
    if (props.mode === 'edit') return t('milvus.editCollection');
    if (props.mode === 'copy') return t('milvus.copyCollection');
    return t('milvus.createCollection');
});

// 是否为编辑模式
const isEditMode = computed(() => props.mode === 'edit');
// 是否为复制模式
const isCopyMode = computed(() => props.mode === 'copy');
// 是否为新增模式
const isCreateMode = computed(() => !props.mode || props.mode === 'create');

// 字段类型分组定义
const fieldTypeGroups = {
    向量: [
        { label: 'FloatVector', value: 101 },
        { label: 'BinaryVector', value: 100 },
        { label: 'Float16Vector', value: 102 },
        { label: 'BFloat16Vector', value: 103 },
        { label: 'SparseFloatVector', value: 104 },
        { label: 'Int8Vector', value: 105 },
    ],
    '数字 & 布尔': [
        { label: 'Int8', value: 2 },
        { label: 'Int16', value: 3 },
        { label: 'Int32', value: 4 },
        { label: 'Int64', value: 5 },
        { label: 'Float', value: 10 },
        { label: 'Double', value: 11 },
        { label: 'Bool', value: 1 },
    ],
    'VarChar & JSON': [
        { label: 'VarChar', value: 21 },
        { label: 'JSON', value: 23 },
    ],
    数组: [{ label: 'Array', value: 22 }],
    其他: [
        { label: 'Geometry', value: 24 },
        { label: 'Timestamptz', value: 26 },
    ],
} as any;

// 数据类型常量映射
const DataTypeMap = {
    Bool: 1,
    Int8: 2,
    Int16: 3,
    Int32: 4,
    Int64: 5,
    Float: 10,
    Double: 11,
    VarChar: 21,
    Array: 22,
    JSON: 23,
    Geometry: 24,
    Timestamptz: 26,
    BinaryVector: 100,
    FloatVector: 101,
    Float16Vector: 102,
    BFloat16Vector: 103,
    SparseFloatVector: 104,
    Int8Vector: 105,
} as const;

// 索引参数模板
const IndexTemplates = {
    // 通用向量索引参数
    vectorCommon: {
        AUTOINDEX: { label: 'AUTOINDEX', group: '自动索引', params: {} },
        FLAT: { label: 'FLAT', group: '内存索引', params: {} },
        IVF_FLAT: { label: 'IVF_FLAT', group: '内存索引', params: { nlist: 1024 } },
        IVF_SQ8: { label: 'IVF_SQ8', group: '内存索引', params: { nlist: 1024 } },
        IVF_PQ: { label: 'IVF_PQ', group: '内存索引', params: { nlist: 1024, m: 0, nbits: 8 } },
        IVF_RABITQ: { label: 'IVF_RABITQ', group: '内存索引', params: { nlist: 1024 } },
        HNSW: { label: 'HNSW', group: '内存索引', params: { M: 16, efConstruction: 200 } },
        HNSW_SQ: { label: 'HNSW_SQ', group: '内存索引', params: { M: 16, efConstruction: 200 } },
        HNSW_PQ: { label: 'HNSW_PQ', group: '内存索引', params: { M: 16, efConstruction: 200, m: 0, nbits: 8 } },
        HNSW_PRQ: { label: 'HNSW_PRQ', group: '内存索引', params: { M: 16, efConstruction: 200, m: 0, nbits: 8 } },
        SCANN: { label: 'SCANN', group: '内存索引', params: { nlist: 1024 } },
        DISKANN: { label: 'DISKANN', group: '磁盘索引', params: {} },
        AISAQ: { label: 'AISAQ', group: '磁盘索引', params: {} },
        GPU_CAGRA: { label: 'GPU_CAGRA', group: 'GPU索引', params: {} },
        GPU_IVF_FLAT: { label: 'GPU_IVF_FLAT', group: 'GPU索引', params: { nlist: 1024 } },
        GPU_IVF_PQ: { label: 'GPU_IVF_PQ', group: 'GPU索引', params: { nlist: 1024, m: 0, nbits: 8 } },
    },
    // BinaryVector 索引
    binaryVector: {
        AUTOINDEX: { label: 'AUTOINDEX', group: '自动索引', params: {} },
        BIN_FLAT: { label: 'BIN_FLAT', group: '内存索引', params: {} },
        BIN_IVF_FLAT: { label: 'BIN_IVF_FLAT', group: '内存索引', params: { nlist: 1024 } },
        IVF_RABITQ: { label: 'IVF_RABITQ', group: '内存索引', params: { nlist: 1024 } },
    },
    // SparseFloatVector 索引
    sparseVector: {
        AUTOINDEX: { label: 'AUTOINDEX', group: '自动索引', params: {} },
        SPARSE_INVERTED_INDEX: { label: 'SPARSE_INVERTED_INDEX', group: '内存索引', params: { drop_ratio_build: 0.0 } },
    },
    // 标量索引
    scalarInverted: {
        AUTOINDEX: { label: 'AUTOINDEX', group: '自动索引', params: {} },
        INVERTED: { label: 'INVERTED', group: '标量索引', params: {} },
    },
    scalarSorted: {
        AUTOINDEX: { label: 'AUTOINDEX', group: '自动索引', params: {} },
        INVERTED: { label: 'INVERTED', group: '标量索引', params: {} },
        STL_SORT: { label: 'STL_SORT', group: '标量索引', params: {} },
    },
    scalarBitmap: {
        AUTOINDEX: { label: 'AUTOINDEX', group: '自动索引', params: {} },
        BITMAP: { label: 'BITMAP', group: '标量索引(推荐)', params: {} },
        INVERTED: { label: 'INVERTED', group: '标量索引', params: {} },
    },
    varcharScalar: {
        AUTOINDEX: { label: 'AUTOINDEX', group: '自动索引', params: {} },
        INVERTED: { label: 'INVERTED', group: '标量索引', params: {} },
        NGRAM: { label: 'NGRAM', group: '标量索引', params: {} },
        BITMAP: { label: 'BITMAP', group: '标量索引', params: {} },
        Trie: { label: 'Trie', group: '标量索引', params: {} },
    },
} as const;

// 辅助函数：创建向量索引配置
const createVectorIndexConfig = (indexes: object, metrics: string[]) => ({
    category: '向量索引',
    indexes,
    metrics,
});

// 辅助函数：创建标量索引配置
const createScalarIndexConfig = (indexes: object) => ({
    category: '标量索引',
    indexes,
    metrics: [],
});

// 索引类型配置 - 根据字段数据类型分类
const IndexConfigByDataType = {
    // 浮点向量类型（共享相同索引配置）
    [DataTypeMap.FloatVector]: createVectorIndexConfig(IndexTemplates.vectorCommon, ['COSINE', 'L2', 'IP']),
    [DataTypeMap.Float16Vector]: createVectorIndexConfig(IndexTemplates.vectorCommon, ['COSINE', 'L2', 'IP']),
    [DataTypeMap.BFloat16Vector]: createVectorIndexConfig(IndexTemplates.vectorCommon, ['COSINE', 'L2', 'IP']),
    [DataTypeMap.Int8Vector]: createVectorIndexConfig(IndexTemplates.vectorCommon, ['COSINE', 'L2', 'IP']),
    // BinaryVector
    [DataTypeMap.BinaryVector]: createVectorIndexConfig(IndexTemplates.binaryVector, ['HAMMING', 'JACCARD', 'MHJACCARD', 'TANIMOTO']),
    // SparseFloatVector
    [DataTypeMap.SparseFloatVector]: createVectorIndexConfig(IndexTemplates.sparseVector, ['IP', 'BM25']),

    // 标量类型
    [DataTypeMap.Int8]: createScalarIndexConfig(IndexTemplates.scalarSorted),
    [DataTypeMap.Int16]: createScalarIndexConfig(IndexTemplates.scalarSorted),
    [DataTypeMap.Int32]: createScalarIndexConfig(IndexTemplates.scalarSorted),
    [DataTypeMap.Int64]: createScalarIndexConfig(IndexTemplates.scalarSorted),
    [DataTypeMap.Float]: createScalarIndexConfig(IndexTemplates.scalarInverted),
    [DataTypeMap.Double]: createScalarIndexConfig(IndexTemplates.scalarInverted),
    [DataTypeMap.Bool]: createScalarIndexConfig(IndexTemplates.scalarBitmap),
    [DataTypeMap.VarChar]: createScalarIndexConfig(IndexTemplates.varcharScalar),
    [DataTypeMap.JSON]: createScalarIndexConfig(IndexTemplates.scalarInverted),
    [DataTypeMap.Array]: createScalarIndexConfig(IndexTemplates.scalarBitmap),
    [DataTypeMap.Geometry]: createScalarIndexConfig(IndexTemplates.scalarInverted),
    [DataTypeMap.Timestamptz]: createScalarIndexConfig(IndexTemplates.scalarSorted),
} as const;

// 向量类型集合
const VectorTypes = new Set<number>([
    DataTypeMap.BinaryVector,
    DataTypeMap.FloatVector,
    DataTypeMap.Float16Vector,
    DataTypeMap.BFloat16Vector,
    DataTypeMap.SparseFloatVector,
    DataTypeMap.Int8Vector,
]);

const form = ref({
    name: '',
    description: '',
    shardsNum: 1,
    consistency_level: 'Bounded',
    fields: [] as any[],
});

// 编辑模式下记录原始 collection 名称
const originalName = ref('');

const rules = {
    name: [Rules.requiredInput('common.name')],
};

// 当前选中的字段
const selectedField = computed(() => {
    if (selectedFieldIndex.value >= 0 && selectedFieldIndex.value < form.value.fields.length) {
        return form.value.fields[selectedFieldIndex.value];
    }
    return null;
});

// 获取主键字段
const getPrimaryKeyField = () => {
    return form.value.fields.find((f) => f.isPrimaryKey);
};

// 判断是否为向量类型
const isVectorType = (type: number) => {
    return VectorTypes.has(type);
};

// 获取字段类型支持的索引配置
const getFieldIndexConfig = (dataType: number) => {
    return (IndexConfigByDataType as any)[dataType] || null;
};

// 判断字段是否支持索引
const supportsIndex = (dataType: number) => {
    return !!getFieldIndexConfig(dataType);
};

// 按组组织索引类型
const getIndexesByGroup = (dataType: number) => {
    const config = getFieldIndexConfig(dataType);
    if (!config) return {};

    const groups: Record<string, any[]> = {};
    for (const [key, index] of Object.entries(config.indexes)) {
        const group = (index as any).group;
        if (!groups[group]) {
            groups[group] = [];
        }
        groups[group].push({ value: key, label: (index as any).label });
    }
    return groups;
};

// 获取度量类型选项
const getMetricOptions = (dataType: number) => {
    const config = getFieldIndexConfig(dataType);
    if (!config) return [];
    return config.metrics || [];
};

// 添加索引到字段
const handleAddIndex = (field: any) => {
    const config = getFieldIndexConfig(field.dataType);
    if (!config) return;

    const firstIndexKey = Object.keys(config.indexes)[0];
    const firstIndex = (config.indexes as any)[firstIndexKey];

    field.indexType = firstIndexKey;
    field.metricType = config.metrics?.[0] || '';
    field.indexParams = { ...firstIndex.params };
};

// 删除字段索引
const handleRemoveIndex = (field: any) => {
    field.indexType = undefined;
    field.metricType = undefined;
    field.indexParams = undefined;
};

// 判断是否为 Geometry 类型
const isGeometryType = (type: number) => {
    return type === DataTypeMap.Geometry;
};

// 判断是否为时间类型
const isClockType = (type: number) => {
    return type === DataTypeMap.Timestamptz;
};

// 判断是否为数组类型
const isArrayType = (type: number) => {
    return type === DataTypeMap.Array;
};

// 判断是否为 VarChar 类型
const isVarCharType = (type: number) => {
    return type === DataTypeMap.VarChar;
};

// 判断是否为 Int64 类型
const isInt64Type = (type: number) => {
    return type === DataTypeMap.Int64;
};

// 判断是否可以作为主键类型（仅 Int64 或 VarChar）
const canBePrimaryKey = (type: number) => {
    return type === DataTypeMap.Int64 || type === DataTypeMap.VarChar;
};

// 获取类型名称
const getTypeName = (type: number): string => {
    const entry = Object.entries(fieldTypeGroups)
        .flatMap(([_, types]) => types)
        .find((t: any) => t.value === type) as any;
    return entry?.label || String(type);
};

// 格式化字段类型显示
const formatFieldType = (field: any) => {
    const typeName = getTypeName(field.dataType);
    if (isVectorType(field.dataType) && field.dim) {
        return `${typeName}(${field.dim})`;
    }
    if (isArrayType(field.dataType) && field.elementType) {
        return `Array<${getTypeName(field.elementType)}>`;
    }
    return typeName;
};

// 选择动态字段
const selectDynamicField = () => {
    selectedFieldIndex.value = -2;
    activeTab.value = 'properties';
};

// 动态字段开关切换
const handleDynamicFieldToggle = (val: boolean) => {
    if (val) {
        // 启用动态字段
        if (!dynamicField.value.indexes || dynamicField.value.indexes.length === 0) {
            dynamicField.value.indexes = [];
        }
    }
};

// 删除动态字段
const handleDeleteDynamicField = () => {
    dynamicFieldEnabled.value = false;
    dynamicField.value = {
        indexes: [],
        selectedIdx: 0,
    };
    selectedFieldIndex.value = -1;
};

// 添加动态字段索引
const handleAddDynamicIndex = () => {
    dynamicField.value.indexes.push({
        indexName: '',
        indexType: 'AUTOINDEX',
        json_path: '$meta["key_name"]',
        json_cast_type: 'varchar',
        json_cast_function: '',
        field_name: '',
    });
    dynamicField.value.selectedIdx = dynamicField.value.indexes.length - 1;
};

// 删除动态字段索引
const handleDeleteDynamicIndex = (index: number) => {
    dynamicField.value.indexes.splice(index, 1);
    if (dynamicField.value.selectedIdx >= dynamicField.value.indexes.length) {
        dynamicField.value.selectedIdx = Math.max(0, dynamicField.value.indexes.length - 1);
    }
};

// 选择字段
const selectField = (index: number) => {
    selectedFieldIndex.value = index;
    activeTab.value = 'properties';
};

// 添加字段
const handleAddField = (dataType: number, name?: string) => {
    const isVector = isVectorType(dataType);
    if (name) {
        // 如果名字存在，则添加后缀_1
        const existField = form.value.fields.find((f) => f.name === name);
        if (existField) {
            name = `${name}_${form.value.fields.filter((f) => f.name === name).length}`;
        }
    }

    if (!dataType) return;
    const fieldIndex = form.value.fields.length + 1;
    const newField = {
        name: name || `field${fieldIndex}`,
        dataType: dataType,
        isPrimaryKey: false,
        autoID: false,
        description: '',
        dim: isVector ? 768 : undefined,
        elementType: isArrayType(dataType) ? DataTypeMap.Int64 : undefined,
        isDynamic: false,
        isPartitionKey: false,
        isClusteringKey: false,
        nullable: false,
        mmap: false,
        defaultValue: '',
        maxLength: isVarCharType(dataType) ? 65535 : undefined,
        maxCapacity: isArrayType(dataType) ? 1024 : undefined,
        // 默认不添加索引
        indexType: undefined,
        metricType: undefined,
        indexParams: undefined,
    };

    // 如果是第一个字段且是 Int64，自动设为主键和自增
    if (form.value.fields.length === 0 && isInt64Type(dataType)) {
        newField.name = 'id';
        newField.isPrimaryKey = true;
        newField.autoID = true;
    }

    form.value.fields.push(newField);
    selectedFieldIndex.value = form.value.fields.length - 1;
    activeTab.value = 'properties';
};

// 删除字段
const handleDeleteField = (index: number) => {
    const field = form.value.fields[index];
    if (field.isPrimaryKey && field.autoID) {
        ElMessage.warning(t('milvus.cannotDeletePrimaryKey'));
        return;
    }
    form.value.fields.splice(index, 1);
    if (selectedFieldIndex.value >= form.value.fields.length) {
        selectedFieldIndex.value = form.value.fields.length - 1;
    }
};
const handleCopyField = (index: number) => {
    const field = form.value.fields[index];

    // 匹配字段名的基础部分和序号(如 field5_1 -> baseName="field5_", num=1)
    const match = field.name.match(/^(.+?)(\d+)$/);

    let baseName: string;
    let startNum: number;

    if (match) {
        baseName = match[1]; // 如 "field5_" 或 "field"
        startNum = parseInt(match[2]); // 如 1 或 2
    } else {
        baseName = field.name + '_';
        startNum = 0;
    }

    // 在所有字段中查找相同基础名称的最大序号
    let maxNum = startNum;
    const escapedBaseName = baseName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`^${escapedBaseName}(\\d+)$`);

    form.value.fields.forEach((f) => {
        const m = f.name.match(regex);
        if (m) {
            const num = parseInt(m[1]);
            if (num > maxNum) {
                maxNum = num;
            }
        }
    });

    // 使用最大序号+1作为新字段名
    const newName = baseName + (maxNum + 1);
    handleAddField(field.dataType, newName);
};

const handleClose = () => {
    visible.value = false;
};

const getFormData = async () => {
    await formRef.value?.validate();
    if (form.value.fields.length === 0) {
        ElMessage.warning(t('milvus.addFieldRequired'));
        return '';
    }

    // 检查是否有主键
    if (!getPrimaryKeyField()) {
        ElMessage.warning(t('milvus.primaryKeyRequired'));
        return '';
    }

    // 检查所有向量字段是否都有索引
    const vectorFieldsWithoutIndex = form.value.fields.filter((field) => isVectorType(field.dataType) && !field.indexType);
    if (vectorFieldsWithoutIndex.length > 0) {
        const fieldNames = vectorFieldsWithoutIndex.map((f) => f.name).join('、');
        ElMessage.warning(t('milvus.vectorFieldIndexRequired', { fields: fieldNames }));
        return '';
    }

    // 构建提交数据
    const fieldsData = form.value.fields.map((field) => {
        const fieldData: any = {
            name: field.name,
            dataType: field.dataType,
            isPrimaryKey: field.isPrimaryKey,
            autoID: field.autoID,
            description: field.description,
            isDynamic: field.isDynamic,
            isPartitionKey: field.isPartitionKey,
            isClusteringKey: field.isClusteringKey,
            nullable: field.nullable,
            mmap: field.mmap,
        };

        // 向量类型添加 dim
        if (isVectorType(field.dataType) && field.dim) {
            fieldData.dim = field.dim;
        }

        // VarChar 类型添加 maxLength
        if (isVarCharType(field.dataType) && field.maxLength) {
            fieldData.maxLength = field.maxLength;
        }

        // Array 类型添加 elementType 和 maxCapacity
        if (isArrayType(field.dataType)) {
            fieldData.elementType = field.elementType;
            fieldData.maxCapacity = field.maxCapacity;
        }

        // 如果设置了 defaultValue
        if (field.defaultValue) {
            fieldData.defaultValue = field.defaultValue;
        }

        // 如果有索引配置
        if (field.indexType && supportsIndex(field.dataType)) {
            fieldData.indexParams = {
                index_type: field.indexType,
                ...(field.metricType && { metric_type: field.metricType }),
                ...field.indexParams,
            };
        }

        // 清空空值
        let obj = {} as any;
        for (const prop in fieldData) {
            if (fieldData[prop] !== undefined && fieldData[prop] !== '' && fieldData[prop] !== null) {
                obj[prop] = fieldData[prop];
            }
        }
        return obj;
    });

    // 添加动态字段
    if (dynamicFieldEnabled.value && dynamicField.value.indexes && dynamicField.value.indexes.length > 0) {
        const dynamicFieldData: any = {
            name: '$meta',
            dataType: DataTypeMap.JSON,
            isDynamic: true,
            indexes: dynamicField.value.indexes.map((idx: any) => {
                const idxData: any = {
                    index_name: idx.indexName || '',
                    index_type: idx.indexType,
                };
                // 动态字段索引特有参数
                if (idx.json_path) idxData.json_path = idx.json_path;
                if (idx.json_cast_type) idxData.json_cast_type = idx.json_cast_type;
                if (idx.json_cast_function) idxData.json_cast_function = idx.json_cast_function;
                if (idx.field_name) idxData.field_name = idx.field_name;
                // 清理空值
                const cleanIdx = {} as any;
                for (const prop in idxData) {
                    if (idxData[prop] !== undefined && idxData[prop] !== '' && idxData[prop] !== null) {
                        cleanIdx[prop] = idxData[prop];
                    }
                }
                return cleanIdx;
            }),
        };
        fieldsData.push(dynamicFieldData);
    }

    const submitData = {
        name: form.value.name,
        description: form.value.description,
        shardsNum: form.value.shardsNum,
        consistency_level: form.value.consistency_level,
        fields: fieldsData,
    };

    return submitData;
};

// 提交表单
const handleSubmit = async () => {
    try {
        const submitData = await getFormData();
        if (!submitData) {
            return;
        }
        loading.value = true;

        if (props.mode === 'edit') {
            // 编辑模式：调用修改接口，提交所有可修改的字段
            // 可修改的字段：newName(重命名), description, consistency_level, fields(新增字段)
            // 不可修改的字段：shardsNum, 已有字段定义
            const alterData: any = {};

            // 1. 重命名：如果名称发生变化
            if (form.value.name !== originalName.value) {
                alterData.newName = form.value.name;
            }

            // 2. 描述
            if (submitData.description !== undefined) {
                alterData.description = submitData.description;
            }

            // 3. 一致性级别
            if (submitData.consistency_level !== undefined) {
                // 将字符串一致性级别转换为数字
                const consistencyMap: Record<string, number> = {
                    'Strong': 0,
                    'Bounded': 1,
                    'Session': 2,
                    'Eventually': 3,
                };
                alterData.consistency_level = consistencyMap[submitData.consistency_level] ?? 1;
            }

            // 4. 新增字段：只提交非只读（新添加）的字段
            const newFields = form.value.fields.filter((f: any) => !f.readonly);
            if (newFields.length > 0) {
                alterData.fields = newFields.map((f: any) => ({
                    name: f.name,
                    data_type: f.dataType,
                    is_primary_key: f.isPrimaryKey || false,
                    auto_id: f.autoID || false,
                    description: f.description || '',
                    dim: f.dim,
                    nullable: f.nullable || false,
                    element_type: f.elementType,
                    is_dynamic: f.isDynamic || false,
                    is_partition_key: f.isPartitionKey || false,
                    is_clustering_key: f.isClusteringKey || false,
                    max_length: f.maxLength,
                    max_capacity: f.maxCapacity,
                    type_params: f.typeParams || {},
                    index_params: f.indexParams || {},
                }));
            }

            await milvusApi.alterCollection(props.milvusId, originalName.value, alterData);
            ElMessage.success(t('milvus.updatedSuccess'));
        } else {
            // 创建/复制模式：调用创建接口
            await milvusApi.createCollection(props.milvusId, submitData);
            ElMessage.success(t('milvus.createdSuccess'));
        }
        visible.value = false;
        emits('success');
    } finally {
        loading.value = false;
    }
};

const handlePreview = async () => {
    const submitData = await getFormData();
    if (!submitData) {
        return;
    }
    MonacoEditorBox({
        content: JSON.stringify(submitData, null, 2),
        title: 'JSON预览',
        language: 'json',
        canChangeLang: false,
        useDrawer: true,
        options: { wordWrap: 'on', tabSize: 2, readOnly: true }, // 自动换行
    });
};

// 重置表单
const resetForm = () => {
    form.value = {
        name: '',
        description: '',
        shardsNum: 1,
        consistency_level: 'Bounded',
        fields: [],
    };
    selectedFieldIndex.value = -1;
    activeTab.value = 'properties';
    dynamicFieldEnabled.value = false;
    dynamicField.value = {
        indexes: [],
        selectedIdx: 0,
    };
    if (formRef.value) {
        formRef.value.clearValidate();
    }
};

// 选择字段类型
const handleSelectFieldType = (dataType: number) => {
    handleAddField(dataType);
    fieldTypePopoverVisible.value = false;
};

// 转换 API 返回的字段数据为表单格式
const transformApiFieldToForm = (f: any) => {
    console.log(f);
    const dim = f.TypeParams?.dim ? parseInt(f.TypeParams.dim) : undefined;
    const maxLength = f.TypeParams?.max_length ? parseInt(f.TypeParams.max_length) : undefined;

    // 解析索引信息
    let indexType = undefined;
    let metricType = undefined;
    let indexParams = undefined;

    // 优先从 IndexParams 中解析（后端 DescribeCollection 返回的结构）
    if (f.IndexParams && Object.keys(f.IndexParams).length > 0) {
        const ip = f.IndexParams;
        indexType = ip.index_type || ip.indexType;
        metricType = ip.metric_type || ip.metricType;

        // 解析 params 字段（可能是 JSON 字符串）
        let paramsObj = {} as any;
        if (ip.params) {
            try {
                paramsObj = typeof ip.params === 'string' ? JSON.parse(ip.params) : ip.params;
            } catch (e) {
                console.warn('Failed to parse index params:', e);
            }
        }

        // 提取其他参数（如 mmap.enabled, nlist 等）
        for (const key in ip) {
            if (['index_type', 'indexType', 'metric_type', 'metricType', 'params'].includes(key)) continue;

            const value = ip[key];
            // 尝试转换为数字
            const numValue = Number(value);
            paramsObj[key] = isNaN(numValue) ? value : numValue;
        }

        if (Object.keys(paramsObj).length > 0) {
            indexParams = paramsObj;
        }
    } else if (f.Indexes && f.Indexes.length > 0) {
        // 兼容旧格式：从 Indexes 数组中获取
        const index = f.Indexes[0];
        indexType = index.IndexType || index.index_type;
        metricType = index.MetricType || index.metric_type;

        const params = index.Params || index.indexParams || {};
        indexParams = {} as any;
        for (const key in params) {
            const value = params[key];
            const numValue = Number(value);
            indexParams[key] = isNaN(numValue) ? value : numValue;
        }
    }

    return {
        name: f.Name || '',
        dataType: f.DataType,
        isPrimaryKey: f.PrimaryKey || false,
        autoID: f.AutoID || false,
        description: f.Description || '',
        isDynamic: f.IsDynamic || false,
        isPartitionKey: f.IsPartitionKey || false,
        isClusteringKey: f.IsClusteringKey || false,
        nullable: f.Nullable || false,
        mmap: false,
        defaultValue: f.DefaultValue || '',
        dim: dim,
        maxLength: maxLength,
        elementType: f.ElementType,
        maxCapacity: undefined,
        readonly: false,
        // 索引信息
        indexType: indexType,
        metricType: metricType,
        indexParams: indexParams,
    };
};

// 监听 visible 变化，初始化表单
watch(visible, (newVal) => {
    if (newVal) {
        resetForm();

        if (props.mode === 'edit' && props.editData) {
            // 编辑模式：填充现有数据
            const schema = props.editData.Schema || props.editData.schema || {};
            // 记录原始名称
            originalName.value = props.editData.Name || props.editData.name || '';
            form.value = {
                name: props.editData.Name || props.editData.name || '',
                description: schema.Description || props.editData.description || '',
                shardsNum: props.editData.ShardNum || 1,
                consistency_level: ['Bounded', 'Strong', 'Session', 'Eventually'][props.editData.ConsistencyLevel] || 'Bounded',
                fields: (schema.Fields || []).map((f: any) => ({
                    ...transformApiFieldToForm(f),
                    readonly: true, // 编辑模式下现有字段只读
                })),
            };
            // 处理动态字段
            if (schema.EnableDynamicField) {
                dynamicFieldEnabled.value = true;
            }
        } else if (props.mode === 'copy' && props.editData) {
            // 复制模式：填充数据，表名加 _copy 后缀
            const schema = props.editData.Schema || props.editData.schema || {};
            const originalName = props.editData.Name || props.editData.name || '';
            const copyName = originalName ? `${originalName}_copy` : '';
            form.value = {
                name: copyName,
                description: schema.Description || props.editData.description || '',
                shardsNum: props.editData.ShardNum || 1,
                consistency_level: ['Bounded', 'Strong', 'Session', 'Eventually'][props.editData.ConsistencyLevel] || 'Bounded',
                fields: (schema.Fields || []).map((f: any) => transformApiFieldToForm(f)),
            };
            // 处理动态字段
            if (schema.EnableDynamicField) {
                dynamicFieldEnabled.value = true;
            }
        } else {
            // 新增模式：默认添加一个主键字段和一个向量字段
            handleAddField(DataTypeMap.Int64);
            handleAddField(DataTypeMap.FloatVector, 'vector');
        }
    }
});
</script>

<style lang="scss" scoped>
:deep(.create-collection-drawer) {
    .el-drawer__header {
        margin-bottom: 20px;
    }
    .el-drawer__body {
        padding-top: 0;
    }
}

.schema-container {
    display: flex;
    gap: 10px;
    height: calc(100vh - 300px);
}

.schema-left {
    flex: 0 0 400px;
    display: flex;
    flex-direction: column;
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
    overflow: hidden;
}

.schema-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color);
    background: var(--el-fill-color-light);
}

.schema-title {
    font-weight: 600;
    font-size: 14px;
}

.field-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
}

.field-item {
    display: flex;
    align-items: center;
    padding: 6px;
    margin-bottom: 8px;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s;
    background: var(--el-fill-color-lighter);
}

.field-item:hover {
    background: var(--el-fill-color);
}

.field-item.active {
    background: var(--el-color-primary-light-9);
    border: 1px solid var(--el-color-primary-light-5);
}

.field-icon {
    margin-right: 12px;
    font-size: 18px;
    color: var(--el-text-color-regular);
}

.field-info {
    flex: 1;
    min-width: 0;
}

.field-name {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
}

.field-name-text {
    font-weight: 500;
    font-size: 14px;
}

.field-type {
    font-size: 12px;
    color: var(--el-text-color-secondary);
}

.field-actions {
    opacity: 0;
    transition: opacity 0.2s;
}

.field-item:hover .field-actions {
    opacity: 1;
}

/* 只读字段样式 */
.field-item.readonly {
    cursor: default;
    opacity: 0.75;
}

.field-item.readonly:hover {
    background: var(--el-fill-color-lighter);
}

/* 动态字段样式 */
.field-item.dynamic-field {
    background: var(--el-color-warning-light-9);
    border: 1px dashed var(--el-color-warning-light-5);
}

.field-item.dynamic-field:hover {
    background: var(--el-color-warning-light-8);
}

.field-item.dynamic-field.active {
    background: var(--el-color-warning-light-9);
    border: 1px solid var(--el-color-warning-light-5);
}

/* 字段索引标签 */
.field-index-tag {
    margin: 0 8px;
    flex-shrink: 0;
}

/* 动态索引列表样式 */
.dynamic-index-list {
    display: flex;
    gap: 16px;
}

.index-list-tabs {
    flex-direction: column;
    gap: 4px;
    max-height: 400px;
    overflow-y: auto;
}

.index-list-tab {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s;
    background: var(--el-fill-color-lighter);
    font-size: 13px;
    margin-top: 8px;
}

.index-list-tab:hover {
    background: var(--el-fill-color);
}

.index-list-tab.active {
    background: var(--el-color-primary-light-9);
    border: 1px solid var(--el-color-primary-light-5);
    color: var(--el-color-primary);
}

.index-list-tab .delete-icon {
    opacity: 0;
    transition: opacity 0.2s;
    cursor: pointer;
    color: var(--el-color-danger);
}

.index-list-tab:hover .delete-icon {
    opacity: 1;
}

.index-empty {
    padding: 20px 0;
}

/* 表单提示 */
.form-tip {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.5;
    margin-top: 4px;
}

.schema-right {
    flex: 1;
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
    padding: 16px;
    overflow-y: auto;
}

.index-config {
    padding: 8px 0;
}

.drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
}

/* 字段类型选择器样式 */
.field-type-popover {
    padding: 0 !important;
}

.field-type-selector {
    display: flex;
    min-height: 300px;
    max-height: 400px;
}

.category-list {
    width: 140px;
    border-right: 1px solid var(--el-border-color-light);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
}

.type-list {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.category-list-header,
.type-list-header {
    padding: 12px 16px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-light);
    border-bottom: 1px solid var(--el-border-color-light);
    font-weight: 500;
}

.category-list-body,
.type-list-body {
    flex: 1;
    overflow-y: auto;
}

.category-item,
.type-item {
    padding: 10px 16px;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 14px;
    color: var(--el-text-color-regular);
}

.category-item:hover,
.type-item:hover {
    background: var(--el-fill-color);
}

.category-item.active {
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-weight: 500;
}

.type-item {
    border-bottom: 1px solid var(--el-border-color-lighter);
}

.type-item:last-child {
    border-bottom: none;
}

.type-item:hover {
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
}
</style>
