<template>
  <div class="gva-table-box">
    <div class="font-medium">手动添加客户</div>
    <el-divider />

    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px" style="max-width: 720px">
      <el-form-item label="公司名称" prop="companyName">
        <el-input v-model.trim="form.companyName" />
      </el-form-item>
      <el-form-item label="邮箱" prop="email">
        <el-input v-model.trim="form.email" />
      </el-form-item>
      <el-form-item label="联系人">
        <el-input v-model.trim="form.contactName" />
      </el-form-item>
      <el-form-item label="职位">
        <el-input v-model.trim="form.title" />
      </el-form-item>
      <el-form-item label="电话">
        <el-input v-model.trim="form.phone" />
      </el-form-item>
      <el-form-item label="国家">
        <el-input v-model.trim="form.country" />
      </el-form-item>
      <el-form-item label="网站">
        <el-input v-model.trim="form.website" />
      </el-form-item>
      <el-form-item label="标签">
        <el-checkbox-group v-model="tagArr">
          <el-checkbox label="潜在" />
          <el-checkbox label="高意向" />
          <el-checkbox label="待联系" />
        </el-checkbox-group>
      </el-form-item>
    </el-form>

    <div class="mt-4 flex gap-2">
      <el-button type="primary" :loading="saving" @click="save(false)">保存</el-button>
      <el-button :loading="saving" @click="save(true)">保存并继续</el-button>
      <el-button @click="cancel">取消</el-button>
    </div>
  </div>
</template>

<script setup>
  import { ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { createContact } from '@/api/contact'

  defineOptions({ name: 'ContactManualAdd' })

  const router = useRouter()
  const formRef = ref(null)
  const saving = ref(false)

  const form = ref({
    companyName: '',
    website: '',
    contactName: '',
    title: '',
    email: '',
    phone: '',
    country: '',
    tags: '',
    status: 'uncontacted',
  })

  const tagArr = ref([])
  watch(tagArr, (v) => {
    form.value.tags = (v || []).join(',')
  }, { deep: true })

  const rules = {
    companyName: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
    email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }],
  }

  const reset = () => {
    form.value = {
      companyName: '',
      website: '',
      contactName: '',
      title: '',
      email: '',
      phone: '',
      country: '',
      tags: '',
      status: 'uncontacted',
    }
    tagArr.value = []
  }

  const save = async (continueAdd) => {
    await formRef.value?.validate()
    saving.value = true
    try {
      const res = await createContact(form.value)
      if (res.code === 0) {
        ElMessage.success('保存成功')
        if (continueAdd) {
          reset()
        } else {
          router.push({ name: 'contactMyList' })
        }
      }
    } finally {
      saving.value = false
    }
  }

  const cancel = () => router.push({ name: 'contactMyList' })
</script>

