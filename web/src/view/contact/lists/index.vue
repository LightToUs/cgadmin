<template>
  <div class="flex gap-3">
    <div class="gva-table-box" style="width: 280px; flex: none">
      <div class="flex items-center justify-between">
        <div class="font-medium">我的列表</div>
        <el-button type="primary" link @click="openCreateList">新建列表</el-button>
      </div>
      <el-divider />

      <el-skeleton v-if="loadingTree" :rows="6" animated />

      <el-tree
        v-else
        :data="treeData"
        node-key="id"
        :props="treeProps"
        :default-expand-all="true"
        highlight-current
        @node-click="handleNodeClick"
      >
        <template #default="{ data }">
          <div class="flex items-center justify-between w-full pr-2">
            <span class="truncate">{{ data.name }}</span>
            <span class="text-xs text-gray-400">{{ data.count }}</span>
          </div>
        </template>
      </el-tree>
    </div>

    <div class="gva-table-box flex-1">
      <div class="flex items-center justify-between flex-wrap gap-2">
        <div class="font-medium">{{ currentListName }}</div>
        <div class="flex items-center gap-2">
          <el-input v-model.trim="keyword" clearable placeholder="搜索公司/联系人/邮箱" style="width: 240px" @keyup.enter="fetchContacts" />
          <el-select v-model="status" clearable placeholder="状态" style="width: 140px" @change="fetchContacts">
            <el-option label="未联系" value="uncontacted" />
            <el-option label="已联系" value="contacted" />
            <el-option label="已回复" value="replied" />
          </el-select>
          <el-select v-model="verified" clearable placeholder="邮箱验证" style="width: 160px" @change="fetchContacts">
            <el-option label="未验证" value="unverified" />
            <el-option label="有效" value="valid" />
            <el-option label="风险" value="risk" />
            <el-option label="无效" value="invalid" />
          </el-select>
          <el-button type="primary" @click="fetchContacts">搜索</el-button>
        </div>
      </div>
      <el-divider />

      <el-table :data="tableData" style="width: 100%" row-key="id" v-loading="loadingTable">
        <el-table-column label="公司名" prop="companyName" min-width="220" />
        <el-table-column label="联系人" prop="contactName" min-width="140" />
        <el-table-column label="邮箱" prop="email" min-width="240" />
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="邮箱验证" width="140">
          <template #default="scope">
            <el-tag :type="verifyType(scope.row.emailVerifyStatus)">{{ verifyText(scope.row.emailVerifyStatus) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button type="primary" link @click="openDetail(scope.row)">详情</el-button>
            <el-button type="primary" link disabled>发送邮件</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="total > 0" class="gva-pagination">
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

    <el-dialog v-model="detailVisible" title="客户详情" width="720px" destroy-on-close>
      <div v-if="detail" class="text-sm text-gray-700 grid grid-cols-2 gap-y-3 gap-x-6">
        <div>公司名：{{ detail.companyName || '-' }}</div>
        <div>网站：{{ detail.website || '-' }}</div>
        <div>联系人：{{ detail.contactName || '-' }}</div>
        <div>职位：{{ detail.title || '-' }}</div>
        <div>邮箱：{{ detail.email || '-' }}</div>
        <div>电话：{{ detail.phone || '-' }}</div>
        <div>国家：{{ detail.country || '-' }}</div>
        <div>状态：{{ statusText(detail.status) }}</div>
        <div>邮箱验证：{{ verifyText(detail.emailVerifyStatus) }}</div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="createVisible" title="新建列表" width="520px" destroy-on-close>
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model.trim="createForm.name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="createList">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getContactList } from '@/api/contact'
  import { createContactList, getContactListTree } from '@/api/contactList'

  defineOptions({ name: 'ContactMyList' })

  const treeProps = { children: 'children', label: 'name' }
  const loadingTree = ref(false)
  const treeData = ref([])
  const currentListId = ref('all')
  const currentListName = ref('全部客户')
  const myListRootId = ref(0)

  const keyword = ref('')
  const status = ref('')
  const verified = ref('')

  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const loadingTable = ref(false)
  const tableData = ref([])

  const statusType = (s) => {
    if (s === 'replied') return 'success'
    if (s === 'contacted') return 'warning'
    return 'info'
  }
  const statusText = (s) => {
    if (s === 'uncontacted') return '未联系'
    if (s === 'contacted') return '已联系'
    if (s === 'replied') return '已回复'
    return s || '-'
  }
  const verifyType = (s) => {
    if (s === 'valid') return 'success'
    if (s === 'risk') return 'warning'
    if (s === 'invalid') return 'danger'
    return 'info'
  }
  const verifyText = (s) => {
    if (s === 'unverified') return '未验证'
    if (s === 'valid') return '有效'
    if (s === 'risk') return '风险'
    if (s === 'invalid') return '无效'
    return s || '-'
  }

  const flatten = (nodes) => {
    const out = []
    const walk = (arr) => {
      for (const n of arr || []) {
        out.push(n)
        if (n.children?.length) walk(n.children)
      }
    }
    walk(nodes)
    return out
  }

  const loadTree = async () => {
    loadingTree.value = true
    try {
      const res = await getContactListTree()
      if (res.code === 0) {
        treeData.value = res.data.tree || []
        const all = flatten(treeData.value)
        const myRoot = all.find(i => i.type === 'group' && i.name === '我的列表')
        myListRootId.value = myRoot?.id || 0
      }
    } finally {
      loadingTree.value = false
    }
  }

  const handleNodeClick = (data) => {
    currentListId.value = data?.id ? String(data.id) : 'all'
    currentListName.value = data?.name || '全部客户'
    page.value = 1
    fetchContacts()
  }

  const fetchContacts = async () => {
    loadingTable.value = true
    try {
      const res = await getContactList({
        page: page.value,
        pageSize: pageSize.value,
        listId: currentListId.value,
        keyword: keyword.value,
        status: status.value,
        verified: verified.value,
      })
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total || 0
      }
    } finally {
      loadingTable.value = false
    }
  }

  const handleCurrentChange = (p) => {
    page.value = p
    fetchContacts()
  }
  const handleSizeChange = (s) => {
    pageSize.value = s
    page.value = 1
    fetchContacts()
  }

  const detailVisible = ref(false)
  const detail = ref(null)
  const openDetail = (row) => {
    detail.value = row
    detailVisible.value = true
  }

  const createVisible = ref(false)
  const createForm = ref({ name: '' })
  const creating = ref(false)
  const openCreateList = () => {
    createForm.value = { name: '' }
    createVisible.value = true
  }
  const createList = async () => {
    if (!createForm.value.name) {
      ElMessage.warning('请输入名称')
      return
    }
    if (!myListRootId.value) {
      ElMessage.warning('列表树未初始化')
      return
    }
    creating.value = true
    try {
      const res = await createContactList({ parentId: myListRootId.value, name: createForm.value.name })
      if (res.code === 0) {
        ElMessage.success('创建成功')
        createVisible.value = false
        await loadTree()
      }
    } finally {
      creating.value = false
    }
  }

  loadTree().then(() => fetchContacts())
</script>

