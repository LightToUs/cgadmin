import service from '@/utils/request'

export const uploadContactImport = (data) => {
  return service({
    url: '/contactImport/upload',
    method: 'post',
    headers: { 'Content-Type': 'multipart/form-data' },
    data
  })
}

export const uploadContactImportGoogleSheet = (data) => {
  return service({
    url: '/contactImport/googleSheet',
    method: 'post',
    data
  })
}

export const suggestContactImportMapping = (data) => {
  return service({
    url: '/contactImport/suggestMapping',
    method: 'post',
    data
  })
}

export const validateContactImport = (data) => {
  return service({
    url: '/contactImport/validate',
    method: 'post',
    data
  })
}

export const startContactImport = (data) => {
  return service({
    url: '/contactImport/start',
    method: 'post',
    data
  })
}

export const getContactImportJob = (params) => {
  return service({
    url: '/contactImport/job',
    method: 'get',
    params
  })
}

export const getContactImportHistory = (params) => {
  return service({
    url: '/contactImport/history',
    method: 'get',
    params
  })
}

export const deleteContactImportJob = (data) => {
  return service({
    url: '/contactImport/delete',
    method: 'post',
    data
  })
}

export const getContactImportErrors = (params) => {
  return service({
    url: '/contactImport/errors',
    method: 'get',
    params
  })
}

export const exportContactImportFailed = (data) => {
  return service({
    url: '/contactImport/exportFailed',
    method: 'post',
    data
  })
}

