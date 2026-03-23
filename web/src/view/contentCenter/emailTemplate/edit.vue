<template>
  <div class="h-full">
    <div class="gva-table-box">
      <div class="flex items-center justify-between flex-wrap gap-2">
        <div class="font-medium">{{ pageTitle }}</div>
        <div class="flex items-center gap-2">
          <el-button icon="back" @click="goBack">返回列表</el-button>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
          <el-button :loading="savingAs" @click="saveAs">另存为</el-button>
          <el-button type="primary" :loading="sending" @click="openTestSend">测试发送</el-button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-2 gap-4">
      <div class="gva-table-box">
        <el-form :model="form" label-width="100px">
          <el-form-item label="模板名称">
            <el-input v-model.trim="form.name" placeholder="例如：首轮开发信-LED行业" />
          </el-form-item>

          <el-form-item label="邮件主题">
            <div class="w-full">
              <el-input
                v-model="form.subject"
                placeholder="例如：Introduction to Our LED Lighting Solutions for {{公司名}}"
                @input="onSubjectInput"
              >
                <template #append>
                  <el-dropdown trigger="click">
                    <el-button>插入变量</el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item v-for="v in variableOptions" :key="v.key" @click="insertToSubject(v.key)">
                          {{ v.label }}
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </template>
              </el-input>
              <div class="text-xs text-gray-500 mt-1 flex justify-end">
                <span :class="subjectLen > 60 ? 'text-red-500' : ''">{{ subjectLen }}/60</span>
              </div>
            </div>
          </el-form-item>

          <el-form-item label="正文编辑">
            <div class="w-full">
              <div class="flex items-center gap-2 mb-2">
                <el-dropdown trigger="click">
                  <el-button>插入变量</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item v-for="v in variableOptions" :key="v.key" @click="insertToEditor(v.key)">
                        {{ v.label }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                <el-select v-model="form.status" style="width: 140px">
                  <el-option label="启用" value="enabled" />
                  <el-option label="停用" value="disabled" />
                </el-select>
              </div>
              <RichEdit v-model="form.bodyHtml" @change="onEditorChange" />
            </div>
          </el-form-item>
        </el-form>
      </div>

      <div class="gva-table-box">
        <div class="flex items-center justify-between gap-2">
          <div class="font-medium">实时预览</div>
          <div class="flex items-center gap-2">
            <el-radio-group v-model="previewMode" size="small">
              <el-radio-button label="mobile">手机</el-radio-button>
              <el-radio-button label="tablet">平板</el-radio-button>
              <el-radio-button label="desktop">桌面</el-radio-button>
            </el-radio-group>
            <el-button icon="refresh" @click="refreshPreview">刷新预览</el-button>
          </div>
        </div>

        <el-divider />

        <div class="text-sm text-gray-700 mb-2">
          <div class="flex items-center justify-between gap-2">
            <div class="truncate">主题：{{ rendered.subject }}</div>
          </div>
        </div>

        <div class="flex justify-center">
          <div class="border border-solid border-gray-200 rounded p-3 bg-white" :style="{ width: previewWidth }">
            <RichView v-model="rendered.bodyHtml" />
          </div>
        </div>

        <el-divider />

        <div class="font-medium mb-2">预览测试数据</div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
          <el-input v-model="previewVars['公司名']" placeholder="公司名" />
          <el-input v-model="previewVars['联系人职位']" placeholder="联系人职位" />
          <el-input v-model="previewVars['联系人姓名']" placeholder="联系人姓名" />
          <el-input v-model="previewVars['国家']" placeholder="国家" />
          <el-input v-model="previewVars['网站']" placeholder="网站" />
        </div>
        <div class="mt-2 flex gap-2">
          <el-button @click="applyPreviewVars">应用</el-button>
          <el-button @click="randomFill">随机填充</el-button>
        </div>

        <el-divider />

        <el-collapse>
          <el-collapse-item title="变量使用说明" name="vars">
            <div class="text-sm text-gray-700 leading-6">
              <div class="font-medium">客户变量</div>
              <div v-for="v in variableOptions.filter(v => v.group === 'customer')" :key="v.key">• {{ v.label }}</div>
              <div class="font-medium mt-2">邮件相关</div>
              <div v-for="v in variableOptions.filter(v => v.group === 'email')" :key="v.key">• {{ v.label }}</div>
              <div class="font-medium mt-2">系统变量</div>
              <div v-for="v in variableOptions.filter(v => v.group === 'system')" :key="v.key">• {{ v.label }}</div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>

    <el-dialog v-model="testSendVisible" title="发送测试邮件" width="620px" destroy-on-close>
      <el-form :model="testSendForm" label-width="100px">
        <el-form-item label="发送到">
          <el-input v-model.trim="testSendForm.to" placeholder="多个邮箱用逗号分隔" />
        </el-form-item>
        <el-form-item label="发件账号">
          <el-select v-model="testSendForm.senderAccountId" class="w-full" clearable placeholder="默认账号">
            <el-option label="默认账号" :value="0" />
            <el-option
              v-for="acc in senderAccounts"
              :key="acc.id"
              :label="`${acc.displayName || acc.email} <${acc.email}>`"
              :value="acc.id"
            />
          </el-select>
        </el-form-item>
        <el-alert type="info" show-icon title="将使用右侧“预览测试数据”进行变量替换并发送测试邮件。" />
      </el-form>
      <template #footer>
        <el-button @click="testSendVisible = false">取消</el-button>
        <el-button type="primary" :loading="sending" @click="doTestSend">发送测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRoute, useRouter } from 'vue-router'
  import RichEdit from '@/components/richtext/rich-edit.vue'
  import RichView from '@/components/richtext/rich-view.vue'
  import { getEmailTemplateDetail, createEmailTemplate, updateEmailTemplate, testSendEmailTemplate } from '@/api/emailTemplate'
  import { getSenderEmailAccountList } from '@/api/senderEmailAccount'

  defineOptions({
    name: 'EmailTemplateEdit'
  })

  const router = useRouter()
  const route = useRoute()

  const id = computed(() => Number(route.params.id || 0))
  const isCreate = computed(() => !id.value)
  const pageTitle = computed(() => (isCreate.value ? '新建邮件模板' : '编辑邮件模板'))

  const form = ref({
    id: 0,
    name: '',
    subject: '',
    bodyHtml: '',
    status: 'enabled',
  })

  const subjectLen = computed(() => (form.value.subject || '').length)

  const variableOptions = [
    { group: 'customer', key: '公司名', label: '{{公司名}}' },
    { group: 'customer', key: '联系人职位', label: '{{联系人职位}}' },
    { group: 'customer', key: '联系人姓名', label: '{{联系人姓名}}' },
    { group: 'customer', key: '国家', label: '{{国家}}' },
    { group: 'customer', key: '网站', label: '{{网站}}' },
    { group: 'email', key: '活动名称', label: '{{活动名称}}' },
    { group: 'email', key: '发送时间', label: '{{发送时间}}' },
    { group: 'system', key: '当前日期', label: '{{当前日期}}' },
    { group: 'system', key: '当前时间', label: '{{当前时间}}' },
    { group: 'system', key: '用户姓名', label: '{{用户姓名}}' },
    { group: 'system', key: '用户邮箱', label: '{{用户邮箱}}' },
    { group: 'system', key: '用户公司', label: '{{用户公司}}' },
    { group: 'system', key: '退订链接', label: '{{退订链接}}' },
  ]

  const previewVars = ref({
    公司名: 'ABC Lighting',
    联系人职位: 'Purchasing Manager',
    联系人姓名: 'Lisa',
    国家: 'United States',
    网站: 'https://www.example.com',
  })

  const previewAppliedVars = ref({ ...previewVars.value })

  const mergeVars = computed(() => ({
    ...previewAppliedVars.value,
  }))

  const renderText = (text, vars) => {
    if (!text) return ''
    return String(text).replace(/\{\{\s*([^}]+?)\s*\}\}/g, (m, key) => {
      const k = String(key || '').trim()
      if (!k) return m
      if (vars && Object.prototype.hasOwnProperty.call(vars, k)) return vars[k]
      return m
    })
  }

  const rendered = computed(() => ({
    subject: renderText(form.value.subject, mergeVars.value),
    bodyHtml: renderText(form.value.bodyHtml, mergeVars.value),
  }))

  const previewMode = ref('desktop')
  const previewWidth = computed(() => {
    if (previewMode.value === 'mobile') return '320px'
    if (previewMode.value === 'tablet') return '768px'
    return '100%'
  })

  const applyPreviewVars = () => {
    previewAppliedVars.value = { ...previewVars.value }
  }

  const randomFill = () => {
    const companies = ['ABC Lighting', 'SunPower LED', 'BrightTech', 'Nova Illumination']
    const titles = ['Purchasing Manager', 'Sales Director', 'CEO', 'Product Manager']
    const countries = ['United States', 'Germany', 'France', 'Canada', 'Australia']
    const pick = (arr) => arr[Math.floor(Math.random() * arr.length)]
    previewVars.value = {
      公司名: pick(companies),
      联系人职位: pick(titles),
      联系人姓名: 'Alex',
      国家: pick(countries),
      网站: 'https://www.example.com',
    }
    applyPreviewVars()
  }

  const onSubjectInput = () => {
    if (subjectLen.value > 120) {
      form.value.subject = form.value.subject.slice(0, 120)
    }
  }

  const insertToSubject = (key) => {
    form.value.subject = `${form.value.subject || ''}{{${key}}}`
  }

  const editorRef = ref(null)
  const onEditorChange = (editor) => {
    editorRef.value = editor
  }

  const insertToEditor = (key) => {
    const editor = editorRef.value
    if (editor?.insertText) {
      editor.insertText(`{{${key}}}`)
      return
    }
    form.value.bodyHtml = `${form.value.bodyHtml || ''}{{${key}}}`
  }

  const refreshPreview = () => {
    previewAppliedVars.value = { ...previewAppliedVars.value }
  }

  const goBack = () => {
    router.push({ name: 'emailTemplate' })
  }

  const saving = ref(false)
  const savingAs = ref(false)

  const validateForm = () => {
    if (!form.value.name) {
      ElMessage.warning('请输入模板名称')
      return false
    }
    if (!form.value.subject) {
      ElMessage.warning('请输入邮件主题')
      return false
    }
    if (!form.value.bodyHtml) {
      ElMessage.warning('请输入正文内容')
      return false
    }
    return true
  }

  const save = async () => {
    if (!validateForm()) return
    saving.value = true
    try {
      if (isCreate.value) {
        const res = await createEmailTemplate({
          name: form.value.name,
          subject: form.value.subject,
          bodyHtml: form.value.bodyHtml,
          bodyText: '',
          status: form.value.status,
        })
        if (res.code === 0) {
          ElMessage.success('创建成功')
          goBack()
        }
        return
      }
      const res = await updateEmailTemplate({
        id: id.value,
        name: form.value.name,
        subject: form.value.subject,
        bodyHtml: form.value.bodyHtml,
        bodyText: '',
        status: form.value.status,
      })
      if (res.code === 0) {
        ElMessage.success('更新成功')
      }
    } finally {
      saving.value = false
    }
  }

  const saveAs = async () => {
    if (!validateForm()) return
    const { value } = await ElMessageBox.prompt('请输入新模板名称', '另存为', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: `${form.value.name}-副本`
    })
    if (!value) return
    savingAs.value = true
    try {
      const res = await createEmailTemplate({
        name: value,
        subject: form.value.subject,
        bodyHtml: form.value.bodyHtml,
        bodyText: '',
        status: form.value.status,
      })
      if (res.code === 0) {
        ElMessage.success('保存成功')
        goBack()
      }
    } finally {
      savingAs.value = false
    }
  }

  const senderAccounts = ref([])
  const loadSenderAccounts = async () => {
    const res = await getSenderEmailAccountList()
    if (res.code === 0) {
      senderAccounts.value = res.data.list || []
    }
  }

  const testSendVisible = ref(false)
  const testSendForm = ref({
    to: '',
    senderAccountId: 0,
  })

  const openTestSend = async () => {
    await loadSenderAccounts()
    testSendVisible.value = true
  }

  const sending = ref(false)
  const doTestSend = async () => {
    if (!id.value) {
      ElMessage.warning('请先保存模板再进行测试发送')
      return
    }
    const toEmails = String(testSendForm.value.to || '')
      .split(',')
      .map(s => s.trim())
      .filter(Boolean)
    if (!toEmails.length) {
      ElMessage.warning('请输入测试邮箱')
      return
    }
    sending.value = true
    try {
      const res = await testSendEmailTemplate({
        templateId: id.value,
        senderAccountId: testSendForm.value.senderAccountId || 0,
        toEmails,
        vars: mergeVars.value,
      })
      if (res.code === 0) {
        ElMessage.success('发送成功')
        testSendVisible.value = false
      }
    } finally {
      sending.value = false
    }
  }

  const init = async () => {
    if (!id.value) return
    const res = await getEmailTemplateDetail({ id: id.value })
    if (res.code === 0) {
      const t = res.data.template
      form.value = {
        id: t.id,
        name: t.name,
        subject: t.subject,
        bodyHtml: t.bodyHtml,
        status: t.status || 'enabled'
      }
    }
  }

  init()
</script>
