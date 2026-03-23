import service from '@/utils/request'

export const getEmailTemplateFolderTree = () => {
  return service({
    url: '/emailTemplateFolder/tree',
    method: 'get'
  })
}

export const createEmailTemplateFolder = (data) => {
  return service({
    url: '/emailTemplateFolder/create',
    method: 'post',
    data
  })
}

export const updateEmailTemplateFolder = (data) => {
  return service({
    url: '/emailTemplateFolder/update',
    method: 'put',
    data
  })
}

export const deleteEmailTemplateFolder = (data) => {
  return service({
    url: '/emailTemplateFolder/delete',
    method: 'delete',
    data
  })
}

