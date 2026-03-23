<template>
  <div class="flex gap-4">
    <div class="w-64">
      <div class="gva-table-box">
        <div class="flex items-center justify-between mb-2">
          <div class="font-medium">文件夹</div>
          <el-button type="primary" link icon="plus" @click="openFolderCreate">新建</el-button>
        </div>
        <el-tree
          :data="folderTree"
          node-key="id"
          :expand-on-click-node="false"
          :props="{ label: 'name', children: 'children' }"
          @node-click="onFolderSelect"
        >
          <template #default="{ data }">
            <div class="flex items-center justify-between w-full">
              <div class="flex items-center gap-2">
                <span v-if="data.color" class="inline-block w-2 h-2 rounded-full" :style="{ background: data.color }" />
                <span class="truncate">{{ data.name }}</span>
              </div>
              <el-dropdown v-if="data.id" trigger="click">
                <el-button type="primary" link icon="more" />
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="openFolderEdit(data)">重命名</el-dropdown-item>
                    <el-dropdown-item divided @click="deleteFolder(data)">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-tree>
        <el-divider />
        <div class="text-sm text-gray-600 cursor-pointer" :class="folderId === null ? 'text-primary' : ''" @click="selectAllFolders">
          所有模板
        </div>
      </div>
    </div>

    <div class="flex-1">
      <div class="gva-search-box">
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <el-radio-group v-model="viewMode" size="small">
              <el-radio-button label="card">网格视图</el-radio-button>
              <el-radio-button label="table">列表视图</el-radio-button>
            </el-radio-group>
          </div>

          <div class="flex items-center gap-2 flex-1 justify-end">
            <el-input v-model="search.keyword" placeholder="搜索模板名称/主题" style="max-width: 360px" @keyup.enter="fetchList" />
            <el-select v-model="search.status" placeholder="状态" clearable style="width: 140px" @change="fetchList">
              <el-option label="启用" value="enabled" />
              <el-option label="停用" value="disabled" />
            </el-select>
            <el-select v-model="search.sortBy" placeholder="排序" clearable style="width: 160px" @change="fetchList">
              <el-option label="最后编辑" value="updatedAt" />
              <el-option label="使用次数" value="usageCount" />
              <el-option label="回复率" value="replyRate" />
              <el-option label="创建时间" value="createdAt" />
            </el-select>
            <el-button type="primary" icon="plus" @click="goCreate">新建模板</el-button>
            <el-dropdown trigger="click">
              <el-button>更多操作</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="openImport">导入模板</el-dropdown-item>
                  <el-dropdown-item divided @click="exportSelected('json')">导出JSON（选中）</el-dropdown-item>
                  <el-dropdown-item @click="exportSelected('html')">导出HTML（选中）</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>

      <div class="gva-table-box">
        <div v-if="viewMode === 'table'" class="gva-btn-list">
          <el-button
            icon="delete"
            style="margin-left: 10px"
            :disabled="!multipleSelection.length"
            @click="batchDelete"
          >删除</el-button>
          <el-button :disabled="!multipleSelection.length" @click="batchStatus('enabled')">启用</el-button>
          <el-button :disabled="!multipleSelection.length" @click="batchStatus('disabled')">停用</el-button>
        </div>

        <el-empty v-if="!tableData.length" description="暂无模板，点击右上角新建模板" />

        <div v-else-if="viewMode === 'card'" class="grid grid-cols-2 xl:grid-cols-3 gap-3">
          <el-card v-for="item in tableData" :key="item.id" shadow="hover">
            <div class="flex items-start justify-between gap-2">
              <div class="font-medium truncate">{{ item.name }}</div>
              <el-tag :type="item.status === 'enabled' ? 'success' : 'info'">
                {{ item.status === 'enabled' ? '启用' : '停用' }}
              </el-tag>
            </div>
            <div class="text-sm text-gray-600 mt-2">
              主题：{{ subjectPreview(item.subject) }}
            </div>
            <div class="text-sm text-gray-600 mt-2 flex items-center justify-between">
              <span>使用次数：{{ item.usageCount || 0 }}次</span>
              <span>回复率：{{ formatRate(item.replyRate) }}</span>
            </div>
            <div class="text-xs text-gray-500 mt-2">
              最后编辑：{{ formatDate(item.updatedAt) }}
            </div>
            <div class="mt-3 flex items-center gap-2">
              <el-button type="primary" link @click="goEdit(item)">编辑</el-button>
              <el-button type="primary" link @click="openPreview(item)">预览</el-button>
              <el-dropdown trigger="click">
                <el-button type="primary" link>更多</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="copyOne(item)">复制</el-dropdown-item>
                    <el-dropdown-item divided @click="deleteOne(item)">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </el-card>
        </div>

        <el-table
          v-else
          ref="multipleTable"
          style="width: 100%"
          tooltip-effect="dark"
          :data="tableData"
          row-key="id"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column align="left" label="模板名称" prop="name" min-width="180" />
          <el-table-column align="left" label="主题预览" min-width="240">
            <template #default="scope">
              <span class="text-gray-700">{{ subjectPreview(scope.row.subject) }}</span>
            </template>
          </el-table-column>
          <el-table-column align="left" label="使用次数" prop="usageCount" width="100" />
          <el-table-column align="left" label="回复率" width="110">
            <template #default="scope">{{ formatRate(scope.row.replyRate) }}</template>
          </el-table-column>
          <el-table-column align="left" label="最后编辑" width="180">
            <template #default="scope">{{ formatDate(scope.row.updatedAt) }}</template>
          </el-table-column>
          <el-table-column align="left" label="状态" width="90">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'enabled' ? 'success' : 'info'">
                {{ scope.row.status === 'enabled' ? '启用' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column align="left" label="操作" width="240">
            <template #default="scope">
              <el-button type="primary" link @click="goEdit(scope.row)">编辑</el-button>
              <el-button type="primary" link @click="openPreview(scope.row)">预览</el-button>
              <el-button type="primary" link @click="copyOne(scope.row)">复制</el-button>
              <el-button type="danger" link @click="deleteOne(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div v-if="viewMode === 'table' && total > 0" class="gva-pagination">
          <el-pagination
            :current-page="page"
            :page-size="pageSize"
            :page-sizes="[10, 30, 50, 100]"
            :total="total"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="handleCurrentChange"
            @size-change="handleSizeChange"
          />
        </div>
      </div>
    </div>

    <el-dialog v-model="previewVisible" :title="previewTitle" width="840px" destroy-on-close>
      <div v-if="previewItem">
        <div class="text-sm text-gray-700">主题：{{ previewItem.subject }}</div>
        <el-divider />
        <div class="border border-solid border-gray-200 rounded p-3">
          <RichView v-model="previewItem.bodyHtml" />
        </div>
        <el-divider />
        <div class="text-sm text-gray-600 flex flex-wrap gap-4">
          <span>使用次数：{{ previewItem.usageCount || 0 }}次</span>
          <span>平均打开率：{{ formatRate(previewItem.openRate) }}</span>
          <span>平均回复率：{{ formatRate(previewItem.replyRate) }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="previewVisible = false">关闭</el-button>
        <el-button type="primary" @click="goEdit(previewItem)">编辑</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="folderDialogVisible" :title="folderDialogTitle" width="520px" destroy-on-close>
      <el-form :model="folderForm" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model.trim="folderForm.name" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-input v-model.trim="folderForm.color" placeholder="#409eff" />
        </el-form-item>
        <el-form-item label="父级">
          <el-select v-model="folderForm.parentId" class="w-full" clearable>
            <el-option label="根目录" :value="0" />
            <el-option
              v-for="opt in folderFlatOptions"
              :key="opt.id"
              :label="opt.name"
              :value="opt.id"
              :disabled="opt.id === folderForm.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="folderDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="folderSaving" @click="saveFolder">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="导入模板" width="560px" destroy-on-close>
      <div class="text-sm text-gray-600 mb-3">
        支持 .json（导入多模板）与 .html（导入单模板）。
      </div>
      <el-upload
        :auto-upload="false"
        :limit="1"
        :on-change="onFileChange"
        :on-remove="onFileRemove"
      >
        <el-button type="primary">选择文件</el-button>
      </el-upload>
      <div class="mt-3">
        <el-radio-group v-model="importMode">
          <el-radio label="create">创建新模板</el-radio>
          <el-radio label="override">同名覆盖</el-radio>
        </el-radio-group>
      </div>
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="doImport">开始导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'
  import RichView from '@/components/richtext/rich-view.vue'
  import { formatDate } from '@/utils/format'
  import {
    batchStatusEmailTemplate,
    copyEmailTemplate,
    deleteEmailTemplate,
    deleteEmailTemplateByIds,
    exportEmailTemplate,
    getEmailTemplateList,
    importEmailTemplate
  } from '@/api/emailTemplate'
  import {
    createEmailTemplateFolder,
    deleteEmailTemplateFolder,
    getEmailTemplateFolderTree,
    updateEmailTemplateFolder
  } from '@/api/emailTemplateFolder'

  defineOptions({
    name: 'EmailTemplateList'
  })

  const router = useRouter()

  const viewMode = ref('card')
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const tableData = ref([])
  const multipleSelection = ref([])

  const folderTree = ref([])
  const folderId = ref(null)

  const search = ref({
    keyword: '',
    status: '',
    sortBy: 'updatedAt',
  })

  const formatRate = (val) => {
    const n = Number(val || 0)
    return `${Math.round(n * 100)}%`
  }

  const subjectPreview = (subject) => {
    if (!subject) return '-'
    const s = String(subject)
    return s.length > 30 ? `${s.slice(0, 30)}...` : s
  }

  const fetchFolders = async () => {
    const res = await getEmailTemplateFolderTree()
    if (res.code === 0) {
      folderTree.value = res.data.tree || []
    }
  }

  const fetchList = async () => {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      keyword: search.value.keyword,
      status: search.value.status,
      sortBy: search.value.sortBy,
      sortDesc: true,
    }
    if (folderId.value !== null && folderId.value !== undefined) {
      params.folderId = folderId.value
    }
    const res = await getEmailTemplateList(params)
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
      page.value = res.data.page || page.value
      pageSize.value = res.data.pageSize || pageSize.value
    }
  }

  const onFolderSelect = (node) => {
    folderId.value = node?.id || null
    page.value = 1
    fetchList()
  }

  const selectAllFolders = () => {
    folderId.value = null
    page.value = 1
    fetchList()
  }

  const handleSelectionChange = (val) => {
    multipleSelection.value = val
  }

  const handleCurrentChange = (val) => {
    page.value = val
    fetchList()
  }

  const handleSizeChange = (val) => {
    pageSize.value = val
    page.value = 1
    fetchList()
  }

  const goCreate = () => {
    router.push({ name: 'emailTemplateEdit', params: { id: 0 } })
  }

  const goEdit = (row) => {
    if (!row) return
    router.push({ name: 'emailTemplateEdit', params: { id: row.id } })
  }

  const deleteOne = async (row) => {
    await ElMessageBox.confirm('确定要删除该模板吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteEmailTemplate({ id: row.id })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchList()
    }
  }

  const batchDelete = async () => {
    await ElMessageBox.confirm('确定要批量删除选中模板吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const ids = multipleSelection.value.map(item => item.id)
    const res = await deleteEmailTemplateByIds({ ids })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchList()
    }
  }

  const batchStatus = async (status) => {
    const ids = multipleSelection.value.map(item => item.id)
    const res = await batchStatusEmailTemplate({ ids, status })
    if (res.code === 0) {
      ElMessage.success('更新成功')
      fetchList()
    }
  }

  const copyOne = async (row) => {
    const res = await copyEmailTemplate({ id: row.id })
    if (res.code === 0) {
      ElMessage.success('复制成功')
      fetchList()
    }
  }

  const exportSelected = async (format) => {
    if (!multipleSelection.value.length) {
      ElMessage.warning('请先在列表视图勾选模板')
      return
    }
    let baseUrl = import.meta.env.VITE_BASE_API
    if (baseUrl === '/') baseUrl = ''
    const ids = multipleSelection.value.map(item => item.id)
    const res = await exportEmailTemplate({ ids, format })
    if (res.code === 0) {
      const url = `${baseUrl}/emailTemplate/downloadByToken?token=${res.data.token}`
      window.open(url, '_blank')
    }
  }

  const importVisible = ref(false)
  const importing = ref(false)
  const importMode = ref('create')
  const importFile = ref(null)

  const openImport = () => {
    importFile.value = null
    importMode.value = 'create'
    importVisible.value = true
  }

  const onFileChange = (file) => {
    importFile.value = file?.raw || null
  }

  const onFileRemove = () => {
    importFile.value = null
  }

  const doImport = async () => {
    if (!importFile.value) {
      ElMessage.warning('请选择文件')
      return
    }
    importing.value = true
    try {
      const formData = new FormData()
      formData.append('file', importFile.value)
      formData.append('mode', importMode.value)
      const res = await importEmailTemplate(formData)
      if (res.code === 0) {
        ElMessage.success(`导入成功：${res.data.count || 0}个`)
        importVisible.value = false
        fetchList()
      }
    } finally {
      importing.value = false
    }
  }

  const previewVisible = ref(false)
  const previewItem = ref(null)
  const previewTitle = computed(() => (previewItem.value ? `模板预览 - ${previewItem.value.name}` : '模板预览'))

  const openPreview = (row) => {
    previewItem.value = row
    previewVisible.value = true
  }

  const folderDialogVisible = ref(false)
  const folderSaving = ref(false)
  const folderMode = ref('create')
  const folderForm = ref({
    id: 0,
    parentId: 0,
    name: '',
    color: '',
    sort: 0,
  })

  const folderDialogTitle = computed(() => (folderMode.value === 'create' ? '新建文件夹' : '编辑文件夹'))

  const flattenFolders = (nodes, acc = [], level = 0) => {
    for (const n of nodes || []) {
      acc.push({ id: n.id, name: `${'—'.repeat(level)}${n.name}`.replace(/^$/, n.name) })
      if (n.children?.length) flattenFolders(n.children, acc, level + 1)
    }
    return acc
  }

  const folderFlatOptions = computed(() => flattenFolders(folderTree.value, []))

  const openFolderCreate = () => {
    folderMode.value = 'create'
    folderForm.value = { id: 0, parentId: 0, name: '', color: '', sort: 0 }
    folderDialogVisible.value = true
  }

  const openFolderEdit = (node) => {
    folderMode.value = 'edit'
    folderForm.value = {
      id: node.id,
      parentId: node.parentId || 0,
      name: node.name,
      color: node.color || '',
      sort: node.sort || 0,
    }
    folderDialogVisible.value = true
  }

  const saveFolder = async () => {
    if (!folderForm.value.name) {
      ElMessage.warning('请输入文件夹名称')
      return
    }
    folderSaving.value = true
    try {
      if (folderMode.value === 'create') {
        const res = await createEmailTemplateFolder({
          parentId: folderForm.value.parentId || 0,
          name: folderForm.value.name,
          color: folderForm.value.color,
          sort: folderForm.value.sort || 0,
        })
        if (res.code === 0) {
          ElMessage.success('创建成功')
          folderDialogVisible.value = false
          fetchFolders()
        }
        return
      }
      const res = await updateEmailTemplateFolder({
        id: folderForm.value.id,
        parentId: folderForm.value.parentId || 0,
        name: folderForm.value.name,
        color: folderForm.value.color,
        sort: folderForm.value.sort || 0,
      })
      if (res.code === 0) {
        ElMessage.success('更新成功')
        folderDialogVisible.value = false
        fetchFolders()
      }
    } finally {
      folderSaving.value = false
    }
  }

  const deleteFolder = async (node) => {
    await ElMessageBox.confirm('确定要删除该文件夹吗？（需确保文件夹下无模板且无子文件夹）', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteEmailTemplateFolder({ id: node.id })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchFolders()
      if (folderId.value === node.id) {
        selectAllFolders()
      }
    }
  }

  fetchFolders()
  fetchList()
</script>
