import service from '@/utils/request'

export const getContactListTree = () => {
  return service({
    url: '/contactList/tree',
    method: 'get'
  })
}

export const createContactList = (data) => {
  return service({
    url: '/contactList/create',
    method: 'post',
    data
  })
}

export const updateContactList = (data) => {
  return service({
    url: '/contactList/update',
    method: 'put',
    data
  })
}

export const deleteContactList = (data) => {
  return service({
    url: '/contactList/delete',
    method: 'delete',
    data
  })
}

