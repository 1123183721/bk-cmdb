<template>
    <div class="template-wrapper" ref="templateWrapper" :style="{ 'padding-top': showFeatureTips ? '10px' : '' }">
        <feature-tips
            :feature-name="'serviceTemplate'"
            :show-tips="showFeatureTips"
            :desc="$t('ServiceManagement[\'功能提示\']')"
            @close-tips="showFeatureTips = false">
        </feature-tips>
        <div class="template-filter clearfix">
            <bk-button class="fl mr10" type="primary" @click="operationTemplate(-1)">{{$t("Common['新建']")}}</bk-button>
            <div class="filter-text fr">
                <cmdb-selector
                    class="fl"
                    :placeholder="$t('ServiceManagement[\'请选择一级分类\']')"
                    :auto-select="false"
                    :allow-clear="true"
                    :list="mainList"
                    v-model="filter['mainClassification']"
                    @on-selected="handleSelect">
                </cmdb-selector>
                <cmdb-selector
                    class="fl"
                    :placeholder="$t('ServiceManagement[\'请选择二级分类\']')"
                    :auto-select="false"
                    :allow-clear="true"
                    :list="secondaryList"
                    v-model="filter['secondaryClassification']"
                    @on-selected="getTableData">
                </cmdb-selector>
                <div class="filter-search fl">
                    <input type="text"
                        class="bk-form-input"
                        :placeholder="$t('ServiceManagement[\'模板名称\']')"
                        v-model.trim="filter.templateName"
                        @keypress.enter="searchByTemplateName">
                    <i class="bk-icon icon-search" @click="searchByTemplateName"></i>
                </div>
            </div>
        </div>
        <cmdb-table class="template-table" ref="table"
            :loading="$loading('get_proc_service_template')"
            :header="table.header"
            :list="table.list"
            :pagination.sync="table.pagination"
            :default-sort="table.defaultSort"
            :wrapper-minus-height="300"
            @handleSortChange="handleSortChange"
            @handleSizeChange="handleSizeChange"
            @handlePageChange="handlePageChange">
            <template slot="last_time" slot-scope="{ item }">
                {{$tools.formatTime(item['last_time'], 'YYYY-MM-DD')}}
            </template>
            <template slot="operation" slot-scope="{ item }">
                <button class="text-primary mr10"
                    @click.stop="operationTemplate(item['id'])">
                    {{$t('Common["编辑"]')}}
                </button>
                <span class="text-primary"
                    style="color: #c4c6cc !important;"
                    v-if="item['service_instance_count']"
                    v-bktooltips.top="$t('ServiceManagement[\'不可删除\']')">
                    {{$t('Common["删除"]')}}
                </span>
                <button class="text-primary"
                    v-else
                    @click.stop="deleteTemplate(item)">
                    {{$t('Common["删除"]')}}
                </button>
            </template>
        </cmdb-table>
    </div>
</template>

<script>
    import { mapGetters, mapActions } from 'vuex'
    import featureTips from '@/components/feature-tips/index'
    export default {
        components: {
            featureTips
        },
        data () {
            return {
                showFeatureTips: false,
                filter: {
                    mainClassification: '',
                    secondaryClassification: '',
                    templateName: ''
                },
                table: {
                    header: [
                        {
                            id: 'name',
                            name: this.$t("ServiceManagement['模板名称']")
                        }, {
                            id: 'service_category',
                            name: this.$t("ServiceManagement['服务分类']")
                        }, {
                            id: 'process_template_count',
                            name: this.$t("ServiceManagement['进程数量']")
                        }, {
                            id: 'service_instance_count',
                            name: this.$t("ServiceManagement['应用数量']")
                        }, {
                            id: 'modifier',
                            name: this.$t("ServiceManagement['修改人']")
                        }, {
                            id: 'last_time',
                            name: this.$t("ServiceManagement['修改时间']")
                        }, {
                            id: 'operation',
                            name: this.$t('Common["操作"]'),
                            sortable: false
                        }
                    ],
                    height: 600,
                    list: [],
                    allList: [],
                    pagination: {
                        current: 1,
                        count: 0,
                        size: 10
                    },
                    defaultSort: '-id',
                    sort: '-id'
                },
                mainList: [],
                secondaryList: [],
                allSecondaryList: [],
                originTemplateData: []
            }
        },
        computed: {
            ...mapGetters(['featureTipsParams']),
            params () {
                const categoryId = this.filter.secondaryClassification ? Number(this.filter.secondaryClassification) : null
                return {
                    service_category_id: categoryId
                }
            }
        },
        async created () {
            this.$store.commit('setHeaderTitle', this.$t("Nav['服务模板']"))
            this.showFeatureTips = this.featureTipsParams['serviceTemplate']
            try {
                await this.getServiceClassification()
                await this.getTableData()
            } catch (e) {
                console.log(e)
            }
        },
        methods: {
            ...mapActions('serviceTemplate', ['searchServiceTemplate', 'deleteServiceTemplate']),
            ...mapActions('serviceClassification', ['searchServiceCategory']),
            async getTableData () {
                const templateData = await this.getTemplateData()
                this.table.allList = templateData.info.map(template => {
                    const result = {
                        process_template_count: template['process_template_count'],
                        service_instance_count: template['service_instance_count'],
                        ...template['service_template']
                    }
                    const secondaryCategory = this.allSecondaryList.find(classification => classification['id'] === result['service_category_id'])
                    const mainCategory = this.mainList.find(classification => secondaryCategory && classification['id'] === secondaryCategory['parent_id'])
                    const secondaryCategoryName = secondaryCategory ? secondaryCategory['name'] : '--'
                    const mainCategoryName = mainCategory ? mainCategory['name'] : '--'
                    result['service_category'] = `${mainCategoryName} / ${secondaryCategoryName}`
                    return result
                })
                this.table.list = this.table.allList
            },
            getTemplateData () {
                return this.searchServiceTemplate({
                    params: this.$injectMetadata(this.params),
                    config: {
                        requestId: 'get_proc_service_template',
                        cancelPrevious: true
                    }
                })
            },
            async getServiceClassification () {
                const res = await this.searchServiceCategory({
                    params: this.$injectMetadata(),
                    config: {
                        requestId: 'get_proc_services_categories'
                    }
                })
                this.classificationList = res.info.map(item => item['category'])
                this.mainList = this.classificationList.filter(classification => !classification['bk_parent_id'])
                this.allSecondaryList = this.classificationList.filter(classification => classification['bk_parent_id'])
            },
            searchByTemplateName () {
                const filterList = this.table.allList.filter(template => template['name'] === this.filter.templateName)
                this.table.list = this.filter.templateName ? filterList : this.table.allList
            },
            handleSelect (id, data) {
                this.secondaryList = this.allSecondaryList.filter(classification => classification['bk_parent_id'] === id && classification['bk_root_id'] === id)
                this.filter.secondaryClassification = ''
            },
            operationTemplate (id) {
                this.$store.commit('setHeaderStatus', {
                    back: true
                })
                this.$router.push({
                    name: 'operationalTemplate',
                    params: {
                        templateId: id
                    }
                })
            },
            deleteTemplate (template) {
                this.$bkInfo({
                    title: this.$t("ServiceManagement['确认删除模版']"),
                    content: this.$tc("ServiceManagement['即将删除服务模版']", name, { name: template.name }),
                    confirmFn: async () => {
                        await this.deleteServiceTemplate({
                            params: {
                                data: this.$injectMetadata({
                                    service_template_id: template.id
                                })
                            },
                            config: {
                                requestId: 'delete_proc_service_template'
                            }
                        })
                        this.getTableData()
                    }
                })
            },
            handleSortChange (sort) {
                this.table.sort = sort
                this.handlePageChange(1)
            },
            handleSizeChange (size) {
                this.table.pagination.size = size
                this.handlePageChange(1)
            },
            handlePageChange (page) {
                this.table.pagination.current = page
                this.getTableData()
            }
        }
    }
</script>

<style lang="scss" scoped>
    .template-wrapper {
        .filter-text {
            .bk-selector {
                width: 180px;
                margin-right: 10px;
            }
            .filter-search {
                width: 200px;
                position: relative;
                .bk-form-input {
                    padding-right: 30px;
                }
                .icon-search {
                    position: absolute;
                    right: 10px;
                    top: 11px;
                    cursor: pointer;
                }
            }
        }
        .template-table {
            margin-top: 15px;
        }
    }
</style>