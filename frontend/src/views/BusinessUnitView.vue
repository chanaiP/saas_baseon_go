<script setup lang="ts">
defineOptions({ name: 'BusinessUnitView' })

import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, ref, watch } from 'vue'

import {
  createBusinessUnitAttrTemplate,
  createBusinessUnitV3,
  deleteBusinessUnitV3,
  deleteBusinessUnitAttrTemplate,
  fetchBusinessUnitActors,
  fetchBusinessUnitAttrTemplate,
  fetchBusinessUnitAttrTemplates,
  fetchBusinessUnitDictionaryTree,
  fetchBusinessUnitRelations,
  fetchBusinessUnitSummary,
  fetchBusinessUnitV3Page,
  matchBusinessUnitAttrTemplate,
  saveBusinessUnitActors,
  saveBusinessUnitRelations,
  updateBusinessUnitAttrTemplate,
  updateBusinessUnitV3,
} from '@/api/businessUnit'
import type {
  BusinessUnitActorRow,
  BusinessUnitAttrField,
  BusinessUnitAttrFieldBody,
  BusinessUnitAttrTemplate,
  BusinessUnitAttrTemplateBody,
  BusinessUnitDictNode,
  BusinessUnitRelationRow,
  BusinessUnitSummaryGroup,
  BusinessUnitSummaryType,
  BusinessUnitV3Body,
  BusinessUnitV3Row,
} from '@/api/businessUnit'
import { fetchOrgTree } from '@/api/organization'
import type { OrgNode } from '@/api/organization'
import { fetchUsers } from '@/api/user'
import type { UserRow } from '@/api/user'
import { usePermissionStore } from '@/stores/permission'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'

const permissionStore = usePermissionStore()

const loading = ref(false)
const summaryLoading = ref(false)
const rows = ref<BusinessUnitV3Row[]>([])
const allUnits = ref<BusinessUnitV3Row[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const keyword = ref('')
const tree = ref<BusinessUnitDictNode[]>([])
const summary = ref<BusinessUnitSummaryType[]>([])
const selectedTypeCode = ref('')
const selectedGroupCode = ref('')
const dialogMode = ref<'create' | 'edit' | null>(null)
const currentEditId = ref<number | null>(null)
const saving = ref(false)
const saveError = ref('')
const templateFields = ref<BusinessUnitAttrField[]>([])
const templateOptions = ref<BusinessUnitAttrTemplate[]>([])
const attrValues = ref<Record<string, unknown>>({})
const actorKeys = ref<string[]>([])
const relationRows = ref<Array<{ target_unit_id: number | null; relation_type_code: string; relation_type_name: string; remark: string }>>([])
const orgTree = ref<OrgNode[]>([])
const users = ref<UserRow[]>([])

const form = ref({
  unit_type_code: '',
  unit_group_code: '',
  name: '',
  code: '',
  parent_id: null as number | null,
  attr_template_id: null as number | null,
  status: 1,
  remark: '',
})

const dialogVisible = computed({
  get: () => dialogMode.value != null,
  set: (value: boolean) => {
    if (!value) dialogMode.value = null
  },
})

const showCreate = computed(() => permissionStore.canUseAction('business_unit:create'))
const canEdit = computed(() => permissionStore.canUseAction('business_unit:edit'))
const canDelete = computed(() => permissionStore.canUseAction('business_unit:delete'))
const canManageActors = computed(() => permissionStore.canUseAction('business_unit:actor_manage') || canEdit.value)
const canManageRelations = computed(() => permissionStore.canUseAction('business_unit:relation_manage') || canEdit.value)
const canManageTemplates = computed(() => permissionStore.canUseAction('business_unit:attr_template_manage'))

const activeType = computed(() => summary.value.find((item) => item.code === selectedTypeCode.value) ?? null)
const activeGroups = computed(() => activeType.value?.groups ?? [])
const activeGroup = computed(() => activeGroups.value.find((item) => item.code === selectedGroupCode.value) ?? null)
const selectedTypeName = computed(() => displayDictName(activeType.value?.name, selectedTypeCode.value) || typeName(selectedTypeCode.value))
const selectedGroupName = computed(() => displayDictName(activeGroup.value?.name, selectedGroupCode.value) || groupName(selectedTypeCode.value, selectedGroupCode.value))
const dialogTitle = computed(() => (dialogMode.value === 'edit' ? '编辑业务单元' : '新增业务单元'))
const typeOptions = computed(() => tree.value)
const groupOptions = computed(() => tree.value.find((item) => item.code === form.value.unit_type_code)?.children ?? [])
const parentOptions = computed(() => allUnits.value.filter((item) => item.id !== currentEditId.value))
const relationTargetOptions = computed(() => allUnits.value.filter((item) => item.id !== currentEditId.value))
const relationTypeOptions = [
  { label: '归属', value: 'belongs_to' },
  { label: '管理', value: 'manages' },
  { label: '服务支持', value: 'supports' },
  { label: '供给', value: 'supplies' },
  { label: '投放', value: 'promotes' },
  { label: '结算', value: 'settles' },
]
const userOptions = computed(() => users.value.map((item) => ({
  label: `${item.name}（${item.employee_no || item.id}）`,
  value: item.id,
})))
const orgSelectOptions = computed(() => orgTree.value.map(toOrgSelectNode))
const selectedOrgKeys = computed({
  get: () => actorKeys.value.filter((key) => key.startsWith('c_') || key.startsWith('d_')),
  set: (keys: string[]) => {
    actorKeys.value = [...keys, ...actorKeys.value.filter((key) => key.startsWith('u_'))]
  },
})
const selectedUserIds = computed({
  get: () => actorKeys.value.filter((key) => key.startsWith('u_')).map((key) => Number(key.slice(2))).filter(Number.isFinite),
  set: (ids: number[]) => {
    const userKeys = ids.map((id) => `u_${id}`)
    actorKeys.value = [...actorKeys.value.filter((key) => key.startsWith('c_') || key.startsWith('d_')), ...userKeys]
  },
})
type TemplateGroupOption = {
  key: string
  typeCode: string
  groupCode: string
  label: string
}

const templateGroupOptions = computed<TemplateGroupOption[]>(() => {
  const selectedTypes = new Set(templateForm.value.unit_type_codes)
  return tree.value
    .filter((type) => selectedTypes.has(type.code))
    .flatMap((type) =>
      type.children.map((group) => ({
        key: templateGroupKey(type.code, group.code),
        typeCode: type.code,
        groupCode: group.code,
        label: `${displayDictName(type.name, type.code)} / ${displayDictName(group.name, group.code)}`,
      })),
    )
})

const detailDrawerVisible = ref(false)
const detailLoading = ref(false)
const detailRow = ref<BusinessUnitV3Row | null>(null)
const detailActors = ref<BusinessUnitActorRow[]>([])
const detailRelations = ref<BusinessUnitRelationRow[]>([])

const templateDrawerVisible = ref(false)
const templateLoading = ref(false)
const templateRows = ref<BusinessUnitAttrTemplate[]>([])
const templateDrawerMode = ref<'list' | 'form'>('list')
const templateDialogMode = ref<'create' | 'edit'>('create')
const templateEditId = ref<number | null>(null)
const templateForm = ref({
  template_name: '',
  unit_type_codes: [] as string[],
  unit_group_keys: [] as string[],
  sort_order: 0,
  status: 'active',
  remark: '',
})

type TemplateFieldForm = {
  field_key: string
  field_label: string
  field_type: BusinessUnitAttrField['field_type']
  required: boolean
  default_value: string
  placeholder: string
  options_json: string
  sort_order: number
  status: string
}

const templateFieldRows = ref<TemplateFieldForm[]>([])

const dictNameAliases: Record<string, string> = {
  ad_account: '投放账号',
  brand: '品牌',
  business_line: '业务线',
  douyin: '抖音',
  jd: '京东',
  jingdong: '京东',
  operation: '运营',
  online_store: '线上店铺',
  project: '项目',
  redbook: '小红书',
  store: '门店',
  warehouse: '仓库',
  xiaohongshu: '小红书',
}

function displayDictName(name: string | null | undefined, code: string | null | undefined) {
  const rawName = (name || '').trim()
  const rawCode = (code || '').trim()
  if (rawName && rawName !== rawCode) return rawName
  return dictNameAliases[rawCode] || rawName || rawCode || '未选择'
}

function typeName(code: string) {
  const item = tree.value.find((node) => node.code === code)
  return displayDictName(item?.name, code)
}

function groupName(typeCode: string, groupCode: string) {
  const type = tree.value.find((item) => item.code === typeCode)
  const item = type?.children?.find((child) => child.code === groupCode)
  return displayDictName(item?.name, groupCode)
}

function templateGroupKey(typeCode: string, groupCode: string) {
  return `${typeCode}::${groupCode}`
}

function parseTemplateGroupKey(key: string) {
  const [typeCode, groupCode] = key.split('::')
  return { typeCode: typeCode || '', groupCode: groupCode || '' }
}

function parseObjectJson(raw: string | null | undefined) {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  } catch {
    return {}
  }
}

function fieldOptions(field: BusinessUnitAttrField) {
  const raw = field.options_json?.trim()
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) {
      return parsed.map((item) =>
        typeof item === 'object' && item
          ? { label: String(item.label ?? item.value ?? ''), value: String(item.value ?? item.label ?? '') }
          : { label: String(item), value: String(item) },
      )
    }
  } catch {
    return []
  }
  return []
}

function defaultAttrValue(field: BusinessUnitAttrField) {
  if (field.default_value != null && field.default_value !== '') {
    if (field.field_type === 'number') return Number(field.default_value)
    if (field.field_type === 'switch') return field.default_value === 'true' || field.default_value === '1'
    return field.default_value
  }
  if (field.field_type === 'switch') return false
  return ''
}

async function loadSummary() {
  summaryLoading.value = true
  try {
    const [dictData, summaryData] = await Promise.all([fetchBusinessUnitDictionaryTree(), fetchBusinessUnitSummary()])
    tree.value = dictData.children ?? []
    summary.value = summaryData.items ?? []
    if (!summary.value.some((item) => item.code === selectedTypeCode.value)) {
      selectedTypeCode.value = summary.value[0]?.code ?? ''
    }
    if (!activeGroups.value.some((item) => item.code === selectedGroupCode.value)) {
      selectedGroupCode.value = activeGroups.value[0]?.code ?? ''
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载业务单元结构失败')
  } finally {
    summaryLoading.value = false
  }
}

async function loadUnits() {
  loading.value = true
  try {
    const data = await fetchBusinessUnitV3Page({
      skip: (page.value - 1) * limit.value,
      limit: limit.value,
      keyword: keyword.value || undefined,
      unit_type_code: selectedTypeCode.value || undefined,
      unit_group_code: selectedGroupCode.value || undefined,
      status: 1,
    })
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载业务单元失败')
  } finally {
    loading.value = false
  }
}

async function loadAllUnits() {
  const data = await fetchBusinessUnitV3Page({ skip: 0, limit: 1000, status: 1 })
  allUnits.value = data.items
}

async function refreshAll() {
  await loadSummary()
  await Promise.all([loadUnits(), loadAllUnits()])
}

function selectType(item: BusinessUnitSummaryType) {
  selectedTypeCode.value = item.code
  selectedGroupCode.value = item.groups[0]?.code ?? ''
  page.value = 1
  void loadUnits()
}

function selectGroup(item: BusinessUnitSummaryGroup) {
  selectedGroupCode.value = item.code
  page.value = 1
  void loadUnits()
}

async function loadTemplate(typeCode: string, groupCode: string) {
  if (!typeCode || !groupCode) {
    templateFields.value = []
    form.value.attr_template_id = null
    return
  }
  try {
    const data = await matchBusinessUnitAttrTemplate({ unit_type_code: typeCode, unit_group_code: groupCode })
    applyTemplateFields(data.fields ?? [])
    form.value.attr_template_id = data.template?.id ?? null
  } catch {
    templateFields.value = []
    form.value.attr_template_id = null
  }
}

function applyTemplateFields(fields: BusinessUnitAttrField[]) {
  templateFields.value = fields
  for (const field of templateFields.value) {
    if (!(field.field_key in attrValues.value)) {
      attrValues.value[field.field_key] = defaultAttrValue(field)
    }
  }
}

async function loadTemplateOptions(typeCode: string, groupCode = '') {
  try {
    const items = await fetchBusinessUnitAttrTemplates({ status: 'active' })
    const normalizedType = typeCode.trim()
    const normalizedGroup = groupCode.trim()
    templateOptions.value = items.filter((item) => {
      if (!normalizedType) return true
      if (item.unit_type_code !== normalizedType) return false
      if (!normalizedGroup) return true
      return !item.unit_group_code || item.unit_group_code === normalizedGroup
    })
  } catch {
    templateOptions.value = []
  }
}

function templateOptionLabel(item: BusinessUnitAttrTemplate) {
  const typeLabel = displayDictName(item.unit_type_name, item.unit_type_code)
  const groupLabel = item.unit_group_code ? displayDictName(item.unit_group_name, item.unit_group_code) : '全部分组'
  return `${item.template_name}（${typeLabel} / ${groupLabel}）`
}

async function applyTemplateById(templateId: number | string | null | undefined) {
  if (!templateId) {
    form.value.attr_template_id = null
    templateFields.value = []
    return
  }
  try {
    const data = await fetchBusinessUnitAttrTemplate(Number(templateId))
    form.value.attr_template_id = data.template.id
    applyTemplateFields(data.fields ?? [])
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载属性模板失败')
  }
}

function walkOrg(nodes: OrgNode[], id: number): OrgNode | null {
  for (const node of nodes) {
    if (node.id === id) return node
    const found = node.children?.length ? walkOrg(node.children, id) : null
    if (found) return found
  }
  return null
}

function toOrgSelectNode(node: OrgNode): { label: string; value: string; children?: Array<{ label: string; value: string; children?: unknown[] }> } {
  const value = `${node.node_type === 'company' ? 'c' : 'd'}_${node.id}`
  const children = node.children?.map(toOrgSelectNode) ?? []
  return {
    label: node.name,
    value,
    children: children.length ? children : undefined,
  }
}

function orgKey(id: number) {
  const node = walkOrg(orgTree.value, id)
  return `${node?.node_type === 'company' ? 'c' : 'd'}_${id}`
}

async function openCreate() {
  dialogMode.value = 'create'
  currentEditId.value = null
  attrValues.value = {}
  actorKeys.value = []
  relationRows.value = []
  form.value = {
    unit_type_code: '',
    unit_group_code: '',
    name: '',
    code: '',
    parent_id: null,
    attr_template_id: null,
    status: 1,
    remark: '',
  }
  await Promise.all([
    loadTemplateOptions(form.value.unit_type_code, form.value.unit_group_code),
    loadTemplate(form.value.unit_type_code, form.value.unit_group_code),
  ])
}

async function openEdit(row: BusinessUnitV3Row) {
  dialogMode.value = 'edit'
  currentEditId.value = row.id
  attrValues.value = parseObjectJson(row.attrs)
  form.value = {
    unit_type_code: row.unit_type_code || '',
    unit_group_code: row.unit_group_code || '',
    name: row.name,
    code: row.code,
    parent_id: row.parent_id,
    attr_template_id: row.attr_template_id,
    status: row.status,
    remark: row.remark ?? '',
  }
  await loadTemplateOptions(form.value.unit_type_code, form.value.unit_group_code)
  if (form.value.attr_template_id) {
    await applyTemplateById(form.value.attr_template_id)
  } else {
    await loadTemplate(form.value.unit_type_code, form.value.unit_group_code)
  }
  try {
    const actors = await fetchBusinessUnitActors(row.id)
    actorKeys.value = actors.map((item: BusinessUnitActorRow) => (item.actor_type === 'user' ? `u_${item.actor_id}` : orgKey(item.actor_id)))
  } catch {
    actorKeys.value = []
  }
  try {
    const relations = await fetchBusinessUnitRelations(row.id)
    relationRows.value = relations.map((item: BusinessUnitRelationRow) => ({
      target_unit_id: item.target_unit_id,
      relation_type_code: item.relation_type_code,
      relation_type_name: item.relation_type_name,
      remark: item.remark ?? '',
    }))
  } catch {
    relationRows.value = []
  }
}

function buildAttrs() {
  const attrs: Record<string, unknown> = {}
  for (const field of templateFields.value) {
    const value = attrValues.value[field.field_key]
    if (field.required && (value == null || value === '')) {
      notifySaveBlocked(`请填写${field.field_label}`)
      return null
    }
    if (value !== '' && value != null) attrs[field.field_key] = value
  }
  return attrs
}

function notifySaveBlocked(message: string) {
  saveError.value = message
  ElMessage.warning(message)
  void ElMessageBox.alert(message, '无法保存', { type: 'warning', confirmButtonText: '知道了' }).catch(() => {})
}

function buildBody(): BusinessUnitV3Body | null {
  if (!form.value.unit_type_code || !form.value.unit_group_code) {
    notifySaveBlocked('请选择业务单元类型和分组')
    return null
  }
  if (!form.value.name.trim() || !form.value.code.trim()) {
    notifySaveBlocked('请填写业务单元名称和编码')
    return null
  }
  const attrs = buildAttrs()
  if (attrs == null) return null
  return {
    unit_type_code: form.value.unit_type_code,
    unit_group_code: form.value.unit_group_code,
    name: form.value.name.trim(),
    code: form.value.code.trim(),
    parent_id: form.value.parent_id,
    attr_template_id: form.value.attr_template_id,
    attrs,
    status: form.value.status,
    remark: form.value.remark.trim() || null,
  }
}

function buildActors() {
  return actorKeys.value
    .map((key) => {
      if (key.startsWith('u_')) {
        return { actor_type: 'user' as const, actor_id: Number(key.slice(2)), role_type: 'owner', include_children: false, status: 'active' }
      }
      if (key.startsWith('c_') || key.startsWith('d_')) {
        return { actor_type: 'org' as const, actor_id: Number(key.slice(2)), role_type: 'owner', include_children: key.startsWith('c_'), status: 'active' }
      }
      return null
    })
    .filter((item): item is { actor_type: 'org' | 'user'; actor_id: number; role_type: string; include_children: boolean; status: string } => !!item && Number.isFinite(item.actor_id))
}

function addRelation() {
  relationRows.value.push({ target_unit_id: null, relation_type_code: '', relation_type_name: '', remark: '' })
}

function removeRelation(index: number) {
  relationRows.value.splice(index, 1)
}

function buildRelations() {
  const incomplete = relationRows.value.find((item) => Boolean(item.target_unit_id) !== Boolean(item.relation_type_code.trim()))
  if (incomplete) {
    notifySaveBlocked('关联业务单元和关系类型需要同时选择')
    return null
  }
  return relationRows.value
    .filter((item) => item.target_unit_id && item.relation_type_code.trim())
    .map((item) => ({
      target_unit_id: Number(item.target_unit_id),
      relation_type_code: item.relation_type_code.trim(),
      relation_type_name: item.relation_type_name.trim() || relationLabel(item.relation_type_code),
      status: 'active',
      remark: item.remark.trim() || null,
    }))
}

function relationLabel(code: string) {
  return relationTypeOptions.find((item) => item.value === code)?.label || code
}

function handleRelationTypeChange(item: { relation_type_code: string; relation_type_name: string }) {
  item.relation_type_name = relationLabel(item.relation_type_code)
}

async function saveDialog() {
  if (saving.value) return
  saveError.value = ''
  const body = buildBody()
  if (!body) return
  const actors = buildActors()
  const relations = buildRelations()
  if (relations == null) return
  saving.value = true
  try {
    const saved = dialogMode.value === 'edit' && currentEditId.value
      ? await updateBusinessUnitV3(currentEditId.value, body)
      : await createBusinessUnitV3(body)
    if (canManageActors.value && (dialogMode.value === 'edit' || actors.length > 0)) {
      await saveBusinessUnitActors(saved.id, actors)
    }
    if (canManageRelations.value && (dialogMode.value === 'edit' || relations.length > 0)) {
      await saveBusinessUnitRelations(saved.id, relations)
    }
    ElMessage.success('业务单元已保存')
    dialogVisible.value = false
    await refreshAll()
  } catch (e) {
    const message = e instanceof Error ? e.message : '保存业务单元失败'
    saveError.value = message
    ElMessage.error(message)
    await ElMessageBox.alert(message, '保存失败', { type: 'error', confirmButtonText: '知道了' }).catch(() => {})
  } finally {
    saving.value = false
  }
}

async function archiveUnit(row: BusinessUnitV3Row) {
  await ElMessageBox.confirm(`确认归档业务单元「${row.name}」？`, '归档确认', { type: 'warning' })
  try {
    await deleteBusinessUnitV3(row.id)
    ElMessage.success('业务单元已归档')
    await refreshAll()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '归档业务单元失败')
  }
}

function unitName(id: number | null | undefined) {
  if (!id) return '-'
  const unit = allUnits.value.find((item) => item.id === id)
  return unit ? `${unit.name}（${unit.code}）` : String(id)
}

function actorName(item: BusinessUnitActorRow) {
  if (item.actor_type === 'user') {
    const user = users.value.find((row) => row.id === item.actor_id)
    return user ? `${user.name}（${user.employee_no || user.id}）` : `用户 ${item.actor_id}`
  }
  const org = walkOrg(orgTree.value, item.actor_id)
  return org ? `${org.name}${item.include_children ? '（含下级）' : ''}` : `组织 ${item.actor_id}`
}

function attrEntries(row: BusinessUnitV3Row | null) {
  const attrs = parseObjectJson(row?.attrs)
  return Object.entries(attrs).filter(([, value]) => value !== '' && value != null)
}

async function openDetail(row: BusinessUnitV3Row) {
  detailRow.value = row
  detailDrawerVisible.value = true
  detailLoading.value = true
  try {
    const [actors, relations] = await Promise.all([
      fetchBusinessUnitActors(row.id),
      fetchBusinessUnitRelations(row.id),
    ])
    detailActors.value = actors
    detailRelations.value = relations
  } catch (e) {
    detailActors.value = []
    detailRelations.value = []
    ElMessage.error(e instanceof Error ? e.message : '加载业务单元详情失败')
  } finally {
    detailLoading.value = false
  }
}

function emptyTemplateField(sortOrder = templateFieldRows.value.length + 1): TemplateFieldForm {
  return {
    field_key: '',
    field_label: '',
    field_type: 'text',
    required: false,
    default_value: '',
    placeholder: '',
    options_json: '',
    sort_order: sortOrder,
    status: 'active',
  }
}

async function loadTemplates() {
  templateLoading.value = true
  try {
    templateRows.value = await fetchBusinessUnitAttrTemplates({
      status: 'active',
    })
  } catch (e) {
    templateRows.value = []
    ElMessage.error(e instanceof Error ? e.message : '加载属性模板失败')
  } finally {
    templateLoading.value = false
  }
}

async function openTemplateDrawer() {
  templateDrawerVisible.value = true
  templateDrawerMode.value = 'list'
  await loadTemplates()
}

function resetTemplateForm() {
  templateForm.value = {
    template_name: '',
    unit_type_codes: [],
    unit_group_keys: [],
    sort_order: 0,
    status: 'active',
    remark: '',
  }
  templateFieldRows.value = [emptyTemplateField(1)]
}

function handleTemplateTypesChange() {
  const selectedTypes = new Set(templateForm.value.unit_type_codes)
  templateForm.value.unit_group_keys = templateForm.value.unit_group_keys.filter((key) => selectedTypes.has(parseTemplateGroupKey(key).typeCode))
}

function openCreateTemplate() {
  templateDialogMode.value = 'create'
  templateEditId.value = null
  resetTemplateForm()
  templateDrawerMode.value = 'form'
}

async function openEditTemplate(row: BusinessUnitAttrTemplate) {
  templateDialogMode.value = 'edit'
  templateEditId.value = row.id
  try {
    const data = await fetchBusinessUnitAttrTemplate(row.id)
    templateForm.value = {
      template_name: data.template.template_name,
      unit_type_codes: [data.template.unit_type_code].filter(Boolean),
      unit_group_keys: data.template.unit_group_code ? [templateGroupKey(data.template.unit_type_code, data.template.unit_group_code)] : [],
      sort_order: data.template.sort_order,
      status: data.template.status,
      remark: data.template.remark ?? '',
    }
    templateFieldRows.value = (data.fields ?? []).map((field) => ({
      field_key: field.field_key,
      field_label: field.field_label,
      field_type: field.field_type,
      required: field.required,
      default_value: field.default_value ?? '',
      placeholder: field.placeholder ?? '',
      options_json: field.options_json ?? '',
      sort_order: field.sort_order,
      status: field.status,
    }))
    if (!templateFieldRows.value.length) templateFieldRows.value = [emptyTemplateField(1)]
    templateDrawerMode.value = 'form'
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载属性模板详情失败')
  }
}

function backToTemplateList() {
  templateDrawerMode.value = 'list'
}

function addTemplateField() {
  templateFieldRows.value.push(emptyTemplateField())
}

function removeTemplateField(index: number) {
  if (templateFieldRows.value.length <= 1) {
    ElMessage.warning('至少保留一个模板字段')
    return
  }
  templateFieldRows.value.splice(index, 1)
}

function parseTemplateFieldOptions(raw: string) {
  const value = raw.trim()
  if (!value) return undefined
  try {
    return JSON.parse(value)
  } catch {
    throw new Error('选项 JSON 格式不正确')
  }
}

function buildTemplateBodies(): BusinessUnitAttrTemplateBody[] | null {
  const selectedTypeCodes = Array.from(new Set(templateForm.value.unit_type_codes.map((item) => item.trim()).filter(Boolean)))
  const selectedGroupKeys = Array.from(new Set(templateForm.value.unit_group_keys.map((item) => item.trim()).filter(Boolean)))
    .map(parseTemplateGroupKey)
    .filter((item) => selectedTypeCodes.includes(item.typeCode) && item.groupCode)
  if (!templateForm.value.template_name.trim() || !selectedTypeCodes.length) {
    ElMessage.warning('请填写模板名称并选择适用业务单元类型')
    return null
  }
  const fields: BusinessUnitAttrFieldBody[] = []
  try {
    for (const [index, field] of templateFieldRows.value.entries()) {
      if (!field.field_key.trim() || !field.field_label.trim()) {
        ElMessage.warning('请填写字段 key 和字段名称')
        return null
      }
      fields.push({
        field_key: field.field_key.trim(),
        field_label: field.field_label.trim(),
        field_type: field.field_type,
        required: field.required,
        default_value: field.default_value.trim() || null,
        placeholder: field.placeholder.trim() || null,
        options_json: field.field_type === 'select' ? parseTemplateFieldOptions(field.options_json) : undefined,
        sort_order: field.sort_order || index + 1,
        status: field.status || 'active',
      })
    }
  } catch (e) {
    ElMessage.warning(e instanceof Error ? e.message : '字段配置不正确')
    return null
  }
  const baseBody = {
    template_name: templateForm.value.template_name.trim(),
    sort_order: templateForm.value.sort_order,
    status: templateForm.value.status,
    remark: templateForm.value.remark.trim() || null,
    fields,
  }
  if (selectedGroupKeys.length) {
    return selectedGroupKeys.map((scope) => ({
      ...baseBody,
      unit_type_code: scope.typeCode,
      unit_group_code: scope.groupCode,
    }))
  }
  return selectedTypeCodes.map((unitTypeCode) => ({
    ...baseBody,
    unit_type_code: unitTypeCode,
    unit_group_code: null,
  }))
}

async function saveTemplateForm() {
  const bodies = buildTemplateBodies()
  if (!bodies?.length) return
  try {
    if (templateDialogMode.value === 'edit' && templateEditId.value) {
      const [firstBody, ...extraBodies] = bodies
      await updateBusinessUnitAttrTemplate(templateEditId.value, firstBody)
      await Promise.all(extraBodies.map((body) => createBusinessUnitAttrTemplate(body)))
    } else {
      await Promise.all(bodies.map((body) => createBusinessUnitAttrTemplate(body)))
    }
    ElMessage.success(bodies.length > 1 ? `属性模板已保存 ${bodies.length} 个适用范围` : '属性模板已保存')
    await loadTemplates()
    await loadTemplateOptions(form.value.unit_type_code, form.value.unit_group_code)
    templateDrawerMode.value = 'list'
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存属性模板失败')
  }
}

async function archiveTemplate(row: BusinessUnitAttrTemplate) {
  try {
    await ElMessageBox.confirm(`确认归档属性模板「${row.template_name}」？`, '归档确认', { type: 'warning' })
    await deleteBusinessUnitAttrTemplate(row.id)
    ElMessage.success('属性模板已归档')
    await loadTemplates()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(e instanceof Error ? e.message : '归档属性模板失败')
    }
  }
}

function handleTemplateMore(command: string, row: BusinessUnitAttrTemplate) {
  if (command === 'archive') void archiveTemplate(row)
}

watch([() => form.value.unit_type_code, () => form.value.unit_group_code], ([typeCode, groupCode]) => {
  if (!dialogVisible.value) return
  void loadTemplateOptions(typeCode, groupCode)
  void loadTemplate(typeCode, groupCode)
})

onMounted(async () => {
  await Promise.all([
    refreshAll(),
    fetchOrgTree().then((data) => { orgTree.value = data }),
    fetchUsers({ skip: 0, limit: 1000 }).then((data) => { users.value = data.items }),
  ])
})
</script>

<template>
  <div class="bu-page">
    <section class="bu-workspace">
      <aside class="bu-type-panel" v-loading="summaryLoading">
        <header class="bu-panel-header">
          <div>
            <p>业务单元类型</p>
            <h1>类型</h1>
          </div>
          <span class="bu-count-pill">{{ summary.length }}</span>
        </header>

        <div class="bu-type-list">
          <button
            v-for="item in summary"
            :key="item.code"
            class="bu-type-card"
            :class="{ active: item.code === selectedTypeCode }"
            type="button"
            @click="selectType(item)"
          >
            <span>
              <strong>{{ displayDictName(item.name, item.code) }}</strong>
              <em>{{ item.unit_count }} 个单元 · {{ item.groups.length }} 个分组</em>
            </span>
            <b>›</b>
          </button>
          <div v-if="!summary.length" class="bu-empty">暂无已登记业务单元</div>
        </div>
      </aside>

      <main class="bu-main">
        <section class="bu-unit-panel">
          <header class="bu-panel-header bu-panel-header-row">
            <div>
              <p>业务单元分组</p>
              <h1>{{ selectedTypeName }}</h1>
              <span v-if="selectedTypeCode" class="bu-header-code">{{ selectedTypeCode }}</span>
            </div>
            <div class="bu-header-actions">
              <el-button v-if="canManageTemplates" class="bu-ghost-btn" @click="openTemplateDrawer">属性模板</el-button>
              <el-button v-if="showCreate" class="bu-primary-btn" type="primary" @click="openCreate">
                新增业务单元
              </el-button>
            </div>
          </header>

          <div class="bu-unit-grid">
            <button
              v-for="item in activeGroups"
              :key="item.code"
              class="bu-unit-card"
              :class="{ active: item.code === selectedGroupCode }"
              type="button"
              @click="selectGroup(item)"
            >
              <strong>{{ displayDictName(item.name, item.code) }}</strong>
              <span>{{ item.code }}</span>
              <em>{{ item.unit_count }} 个单元</em>
            </button>
            <div v-if="!activeGroups.length" class="bu-empty">当前类型下暂无业务单元分组</div>
          </div>
        </section>

        <section class="bu-list-panel">
          <header class="bu-list-header">
            <div>
              <p>{{ selectedTypeName }} / {{ selectedGroupName }}</p>
              <h2>业务单元</h2>
            </div>
            <div class="bu-search">
              <el-input v-model="keyword" clearable placeholder="搜索名称/编码" @keyup.enter="loadUnits" />
              <el-button @click="loadUnits">查询</el-button>
            </div>
          </header>

          <el-table v-loading="loading" :data="rows" class="bu-table" table-layout="fixed">
            <el-table-column label="名称" prop="name" min-width="150" />
            <el-table-column label="编码" prop="code" min-width="140" />
            <el-table-column label="分组" min-width="120">
              <template #default="{ row }">{{ displayDictName(row.unit_group_name, row.unit_group_code) }}</template>
            </el-table-column>
            <el-table-column label="父级" min-width="120">
              <template #default="{ row }">{{ row.parent_id || '-' }}</template>
            </el-table-column>
            <el-table-column label="状态" min-width="90">
              <template #default="{ row }">
                <el-tag size="small" :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="230" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="openDetail(row)">详情</el-button>
                <el-button v-if="canEdit" size="small" @click="openEdit(row)">编辑</el-button>
                <el-button v-if="canDelete" size="small" type="danger" plain @click="archiveUnit(row)">归档</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div v-if="total > limit" class="bu-pager">
            <el-pagination
              v-model:current-page="page"
              :page-size="limit"
              :total="total"
              layout="total, prev, pager, next"
              @current-change="loadUnits"
            />
          </div>
        </section>
      </main>
    </section>

    <el-drawer
      v-model="detailDrawerVisible"
      class="bu-side-drawer"
      :title="detailRow?.name || '业务单元详情'"
      direction="rtl"
      size="min(640px, 94vw)"
      append-to-body
    >
      <div v-if="detailRow" v-loading="detailLoading" class="bu-drawer-body">
        <section class="bu-drawer-section">
          <h3>基础信息</h3>
          <div class="bu-info-grid">
            <div class="bu-info-item"><span>名称</span><strong>{{ detailRow.name }}</strong></div>
            <div class="bu-info-item"><span>编码</span><strong>{{ detailRow.code }}</strong></div>
            <div class="bu-info-item"><span>类型</span><strong>{{ displayDictName(detailRow.unit_type_name, detailRow.unit_type_code) }}</strong></div>
            <div class="bu-info-item"><span>分组</span><strong>{{ displayDictName(detailRow.unit_group_name, detailRow.unit_group_code) }}</strong></div>
            <div class="bu-info-item"><span>父级</span><strong>{{ unitName(detailRow.parent_id) }}</strong></div>
            <div class="bu-info-item"><span>状态</span><strong>{{ detailRow.status === 1 ? '启用' : '停用' }}</strong></div>
          </div>
          <p v-if="detailRow.remark" class="bu-detail-remark">{{ detailRow.remark }}</p>
        </section>

        <section class="bu-drawer-section">
          <h3>负责组织 / 负责人</h3>
          <div v-if="detailActors.length" class="bu-chip-list">
            <span v-for="item in detailActors" :key="item.id" class="bu-chip">{{ actorName(item) }}</span>
          </div>
          <div v-else class="bu-empty compact">暂未绑定负责组织或负责人</div>
        </section>

        <section class="bu-drawer-section">
          <h3>关联业务单元</h3>
          <div v-if="detailRelations.length" class="bu-detail-list">
            <div v-for="item in detailRelations" :key="item.id" class="bu-detail-list-item">
              <strong>{{ item.relation_type_name }}</strong>
              <span>{{ unitName(item.target_unit_id) }}</span>
              <em v-if="item.remark">{{ item.remark }}</em>
            </div>
          </div>
          <div v-else class="bu-empty compact">暂未维护关联业务单元</div>
        </section>

        <section class="bu-drawer-section">
          <h3>自定义属性</h3>
          <div v-if="attrEntries(detailRow).length" class="bu-info-grid">
            <div v-for="[key, value] in attrEntries(detailRow)" :key="key" class="bu-info-item">
              <span>{{ key }}</span>
              <strong>{{ value }}</strong>
            </div>
          </div>
          <div v-else class="bu-empty compact">暂无自定义属性</div>
        </section>
      </div>
    </el-drawer>

    <el-drawer
      v-model="templateDrawerVisible"
      class="bu-side-drawer"
      title="属性模板"
      direction="rtl"
      size="min(780px, 96vw)"
      append-to-body
    >
      <div class="bu-drawer-body">
        <div v-if="templateDrawerMode === 'list'" class="bu-template-list-view">
          <div class="bu-drawer-toolbar">
            <span>按类型或分组配置字段模板，新增业务单元时自动加载。</span>
            <el-button v-if="canManageTemplates" class="bu-primary-btn small" type="primary" @click="openCreateTemplate">新增模板</el-button>
          </div>
          <el-table v-loading="templateLoading" :data="templateRows" class="bu-table" table-layout="fixed">
            <el-table-column label="模板名称" prop="template_name" min-width="150" />
            <el-table-column label="适用类型" min-width="110">
              <template #default="{ row }">{{ displayDictName(row.unit_type_name, row.unit_type_code) }}</template>
            </el-table-column>
            <el-table-column label="适用分组" min-width="110">
              <template #default="{ row }">{{ row.unit_group_code ? displayDictName(row.unit_group_name, row.unit_group_code) : '全部分组' }}</template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag size="small" :type="row.status === 'active' ? 'success' : 'info'">{{ row.status === 'active' ? '启用' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="openEditTemplate(row)">编辑</el-button>
                <el-dropdown trigger="click" @command="(command: string) => handleTemplateMore(command, row)">
                  <el-button class="bu-more-btn" size="small">...</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="archive">归档</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div v-else class="bu-template-form-view">
          <div class="bu-drawer-toolbar">
            <div>
              <strong>{{ templateDialogMode === 'edit' ? '编辑属性模板' : '新增属性模板' }}</strong>
              <span>字段会在新增业务单元时按类型和分组自动加载。</span>
            </div>
            <el-button class="bu-ghost-btn small" @click="backToTemplateList">返回列表</el-button>
          </div>

          <el-form label-position="top" class="bu-form">
            <div class="bu-form-section">
              <h3>模板范围</h3>
              <div class="bu-drawer-form-grid">
                <el-form-item label="模板名称">
                  <el-input v-model="templateForm.template_name" placeholder="例如：门店通用属性" />
                </el-form-item>
                <el-form-item label="适用业务单元类型">
                  <el-select
                    v-model="templateForm.unit_type_codes"
                    placeholder="请选择适用类型"
                    multiple
                    collapse-tags
                    collapse-tags-tooltip
                    filterable
                    @change="handleTemplateTypesChange"
                  >
                    <el-option v-for="item in typeOptions" :key="item.code" :label="displayDictName(item.name, item.code)" :value="item.code" />
                  </el-select>
                </el-form-item>
                <el-form-item label="适用业务单元分组">
                  <el-select
                    v-model="templateForm.unit_group_keys"
                    placeholder="不选则适用所选类型下全部分组"
                    multiple
                    collapse-tags
                    collapse-tags-tooltip
                    clearable
                    filterable
                  >
                    <el-option v-for="item in templateGroupOptions" :key="item.key" :label="item.label" :value="item.key" />
                  </el-select>
                </el-form-item>
                <el-form-item label="状态">
                  <el-select v-model="templateForm.status">
                    <el-option label="启用" value="active" />
                    <el-option label="停用" value="disabled" />
                  </el-select>
                </el-form-item>
              </div>
              <el-form-item label="备注">
                <el-input v-model="templateForm.remark" type="textarea" :rows="2" placeholder="可为空" />
              </el-form-item>
            </div>

            <div class="bu-form-section">
              <div class="bu-section-title-row">
                <h3>字段配置</h3>
                <el-button size="small" @click="addTemplateField">新增字段</el-button>
              </div>
              <div class="bu-template-field-list">
                <div
                  v-for="(field, index) in templateFieldRows"
                  :key="index"
                  class="bu-template-field-row drawer"
                  :class="{ 'has-options': field.field_type === 'select' }"
                >
                  <el-input v-model="field.field_label" class="field-label" placeholder="字段名称" />
                  <el-input v-model="field.field_key" class="field-key" placeholder="字段 key" />
                  <el-select v-model="field.field_type" class="field-type" placeholder="类型">
                    <el-option label="文本" value="text" />
                    <el-option label="数字" value="number" />
                    <el-option label="日期" value="date" />
                    <el-option label="下拉" value="select" />
                    <el-option label="多行文本" value="textarea" />
                    <el-option label="开关" value="switch" />
                  </el-select>
                  <el-input v-model="field.default_value" class="field-default" placeholder="默认值" />
                  <el-input v-model="field.placeholder" class="field-placeholder" placeholder="占位提示" />
                  <div class="field-required">
                    <span>必填</span>
                    <el-switch v-model="field.required" />
                  </div>
                  <el-button class="field-remove" text type="danger" @click="removeTemplateField(index)">移除</el-button>
                  <el-form-item v-if="field.field_type === 'select'" class="bu-template-options" label="">
                    <el-input
                      v-model="field.options_json"
                      type="textarea"
                      :rows="2"
                      placeholder='例如 ["直营","加盟"]'
                    />
                  </el-form-item>
                </div>
              </div>
            </div>
          </el-form>

          <div class="bu-drawer-footer">
            <el-button @click="backToTemplateList">取消</el-button>
            <el-button type="primary" @click="saveTemplateForm">确认</el-button>
          </div>
        </div>
      </div>
    </el-drawer>

    <NeuroAgentDialog v-model="dialogVisible" :title="dialogTitle" width="980px">
      <el-form label-position="top" class="bu-form">
        <div class="bu-form-section">
          <h3>基础信息</h3>
          <div class="bu-form-grid">
            <el-form-item label="业务单元类型">
              <el-select v-model="form.unit_type_code" placeholder="请选择类型" filterable>
                <el-option v-for="item in typeOptions" :key="item.code" :label="displayDictName(item.name, item.code)" :value="item.code" />
              </el-select>
            </el-form-item>
            <el-form-item label="业务单元分组">
              <el-select v-model="form.unit_group_code" placeholder="请选择分组" filterable>
                <el-option v-for="item in groupOptions" :key="item.code" :label="displayDictName(item.name, item.code)" :value="item.code" />
              </el-select>
            </el-form-item>
            <el-form-item label="业务单元名称">
              <el-input v-model="form.name" placeholder="例如：抖音南京旗舰店" />
            </el-form-item>
            <el-form-item label="业务单元编码">
              <el-input v-model="form.code" placeholder="例如：douyin_nanjing_001" />
            </el-form-item>
            <el-form-item label="父级业务单元">
              <el-select v-model="form.parent_id" placeholder="可为空" clearable filterable>
                <el-option v-for="item in parentOptions" :key="item.id" :label="item.name" :value="item.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="form.status">
                <el-option label="启用" :value="1" />
                <el-option label="停用" :value="0" />
              </el-select>
            </el-form-item>
          </div>
          <el-form-item label="备注">
            <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="可为空" />
          </el-form-item>
        </div>

        <div class="bu-form-section">
          <h3>自定义属性</h3>
          <el-form-item label="属性模板">
            <el-select
              v-model="form.attr_template_id"
              placeholder="自动匹配，可手动选择"
              clearable
              filterable
              @change="applyTemplateById"
            >
              <el-option
                v-for="item in templateOptions"
                :key="item.id"
                :label="templateOptionLabel(item)"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <div v-if="templateFields.length" class="bu-form-grid">
            <el-form-item v-for="field in templateFields" :key="field.id" :label="field.field_label" :required="field.required">
              <el-input-number v-if="field.field_type === 'number'" v-model="attrValues[field.field_key]" controls-position="right" />
              <el-date-picker v-else-if="field.field_type === 'date'" v-model="attrValues[field.field_key]" type="date" value-format="YYYY-MM-DD" />
              <el-select v-else-if="field.field_type === 'select'" v-model="attrValues[field.field_key]" :placeholder="field.placeholder || '请选择'">
                <el-option v-for="option in fieldOptions(field)" :key="option.value" :label="option.label" :value="option.value" />
              </el-select>
              <el-switch v-else-if="field.field_type === 'switch'" v-model="attrValues[field.field_key]" />
              <el-input v-else-if="field.field_type === 'textarea'" v-model="attrValues[field.field_key]" type="textarea" :rows="3" :placeholder="field.placeholder || ''" />
              <el-input v-else v-model="attrValues[field.field_key]" :placeholder="field.placeholder || ''" />
            </el-form-item>
          </div>
          <div v-else class="bu-empty compact">当前类型和分组未配置属性模板</div>
        </div>

        <div v-if="canManageActors" class="bu-form-section">
          <h3>负责组织 / 负责人</h3>
          <div class="bu-actor-grid">
            <el-form-item label="负责组织">
              <el-tree-select
                v-model="selectedOrgKeys"
                :data="orgSelectOptions"
                multiple
                check-strictly
                collapse-tags
                collapse-tags-tooltip
                clearable
                filterable
                placeholder="可多选组织，可为空"
              />
            </el-form-item>
            <el-form-item label="负责人">
              <el-select
                v-model="selectedUserIds"
                multiple
                collapse-tags
                collapse-tags-tooltip
                clearable
                filterable
                placeholder="可多选人员，可为空"
              >
                <el-option v-for="item in userOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </div>
        </div>

        <div v-if="canManageRelations" class="bu-form-section">
          <div class="bu-section-title-row">
            <h3>关联业务单元</h3>
            <el-button size="small" @click="addRelation">新增关联</el-button>
          </div>
          <div v-if="relationRows.length" class="bu-relation-list">
            <div v-for="(item, index) in relationRows" :key="index" class="bu-relation-row">
              <el-select v-model="item.target_unit_id" placeholder="选择业务单元" filterable>
                <el-option
                  v-for="unit in relationTargetOptions"
                  :key="unit.id"
                  :label="`${unit.name}（${unit.code}）`"
                  :value="unit.id"
                />
              </el-select>
              <el-select v-model="item.relation_type_code" placeholder="关系类型" filterable @change="handleRelationTypeChange(item)">
                <el-option v-for="option in relationTypeOptions" :key="option.value" :label="option.label" :value="option.value" />
              </el-select>
              <el-input v-model="item.remark" placeholder="备注" />
              <el-button text type="danger" @click="removeRelation(index)">移除</el-button>
            </div>
          </div>
          <div v-else class="bu-empty compact">暂未维护关联关系</div>
        </div>
      </el-form>

      <template #footer>
        <div v-if="saveError" class="bu-save-error">{{ saveError }}</div>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button native-type="button" type="primary" :loading="saving" :disabled="saving" @click.stop.prevent="saveDialog">确认</el-button>
      </template>
    </NeuroAgentDialog>

  </div>
</template>

<style scoped>
.bu-page {
  box-sizing: border-box;
  min-height: 100%;
  padding: 14px 0 0;
  overflow: hidden;
  color: #f5f7fb;
  background: #111827;
}

.bu-workspace {
  display: grid;
  grid-template-columns: 400px minmax(0, 1fr);
  gap: 20px;
  width: 100%;
  max-width: 100%;
  height: min(640px, calc(100vh - 262px));
  min-height: 400px;
}

.bu-type-panel,
.bu-unit-panel,
.bu-list-panel {
  overflow: hidden;
  border: 1px solid rgba(125, 155, 190, 0.22);
  border-radius: 18px;
  background:
    linear-gradient(180deg, rgba(20, 37, 52, 0.92), rgba(16, 25, 39, 0.96)),
    #111827;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.18);
}

.bu-type-panel {
  min-height: 0;
}

.bu-main {
  display: grid;
  grid-template-rows: 260px auto;
  gap: 20px;
  min-width: 0;
}

.bu-unit-panel {
  display: grid;
  grid-template-rows: auto 1fr;
  min-height: 0;
}

.bu-panel-header,
.bu-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 22px 26px;
  border-bottom: 1px solid rgba(125, 155, 190, 0.18);
}

.bu-panel-header p,
.bu-list-header p {
  margin: 0 0 6px;
  color: #16f2dc;
  font-size: 13px;
  font-weight: 800;
}

.bu-panel-header h1,
.bu-list-header h2 {
  margin: 0;
  font-size: 22px;
  line-height: 1.15;
  letter-spacing: 0;
}

.bu-list-header h2 {
  font-size: 24px;
}

.bu-header-code {
  display: block;
  margin-top: 6px;
  color: #91a2ba;
  font-size: 12px;
  font-weight: 700;
}

.bu-count-pill {
  display: grid;
  width: 52px;
  height: 52px;
  place-items: center;
  border-radius: 999px;
  color: #16f2dc;
  background: rgba(22, 242, 220, 0.16);
  font-size: 22px;
  font-weight: 900;
}

.bu-header-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
}

.bu-type-list {
  display: grid;
  gap: 14px;
  padding: 18px 24px;
}

.bu-type-card {
  width: 100%;
  height: 90px;
  padding: 16px 18px;
  border: 1px solid rgba(140, 164, 194, 0.26);
  border-radius: 14px;
  color: #f5f7fb;
  background: rgba(15, 26, 40, 0.68);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  text-align: left;
}

.bu-type-card.active,
.bu-unit-card.active {
  border-color: #12e8d1;
  background: rgba(22, 242, 220, 0.12);
  box-shadow: inset 0 0 0 1px rgba(18, 232, 209, 0.7), 0 12px 38px rgba(18, 232, 209, 0.08);
}

.bu-type-card strong,
.bu-unit-card strong {
  display: block;
  overflow: hidden;
  color: #f9fbff;
  font-size: 16px;
  font-weight: 900;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bu-type-card em,
.bu-unit-card span,
.bu-unit-card em {
  display: block;
  margin-top: 7px;
  color: #9aabc3;
  font-size: 12px;
  font-style: normal;
  font-weight: 700;
}

.bu-type-card b {
  color: #9aabc3;
  font-size: 24px;
}

.bu-unit-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  align-content: flex-start;
  min-height: 0;
  padding: 20px 28px 24px;
  overflow: hidden;
}

.bu-unit-card {
  position: relative;
  width: 180px;
  height: 90px;
  padding: 15px 16px;
  border: 1px solid rgba(140, 164, 194, 0.26);
  border-radius: 14px;
  color: #f5f7fb;
  background: rgba(15, 26, 40, 0.68);
  cursor: pointer;
  text-align: left;
}

.bu-unit-card em {
  position: absolute;
  right: 14px;
  bottom: 12px;
  color: #16f2dc;
}

.bu-primary-btn {
  min-width: 126px;
  height: 40px;
  border: 0;
  color: #06131f;
  background: linear-gradient(135deg, #43efe4, #11d3b7);
  font-weight: 900;
  font-size: 14px;
  box-shadow: 0 12px 32px rgba(17, 211, 183, 0.32);
}

.bu-primary-btn.small {
  min-width: 104px;
  height: 36px;
}

.bu-ghost-btn {
  min-width: 96px;
  height: 40px;
  border-color: rgba(22, 242, 220, 0.45);
  color: #16f2dc;
  background: rgba(15, 26, 40, 0.72);
  font-weight: 800;
}

.bu-ghost-btn.small {
  min-width: 84px;
  height: 34px;
}

.bu-more-btn {
  width: 34px;
  min-width: 34px;
  padding: 0;
  margin-left: 8px;
  font-weight: 900;
}

.bu-list-header {
  align-items: flex-end;
  flex-wrap: wrap;
  gap: 14px;
  padding-right: 28px;
}

.bu-search {
  display: flex;
  gap: 10px;
  width: min(360px, 100%);
}

.bu-table {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: rgba(22, 242, 220, 0.1);
  --el-table-header-text-color: #eaf4ff;
  --el-table-text-color: #f5f7fb;
  --el-table-row-hover-bg-color: rgba(22, 242, 220, 0.06);
  --el-table-border-color: rgba(125, 155, 190, 0.18);
  font-size: 13px;
}

.bu-table :deep(.el-table__cell) {
  padding: 12px 0;
}

.bu-pager {
  display: flex;
  justify-content: flex-end;
  padding: 14px 18px;
}

.bu-form {
  display: grid;
  gap: 18px;
}

.bu-form-section {
  padding: 18px;
  border: 1px solid rgba(125, 155, 190, 0.18);
  border-radius: 14px;
  background: rgba(15, 26, 40, 0.58);
}

.bu-form-section h3 {
  margin: 0 0 14px;
  color: #16f2dc;
  font-size: 15px;
  font-weight: 900;
}

.bu-section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.bu-section-title-row h3 {
  margin: 0;
}

.bu-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 18px;
}

.bu-actor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 18px;
}

.bu-empty {
  padding: 22px;
  color: #96a6bd;
  font-size: 14px;
}

.bu-empty.compact {
  padding: 8px 0;
}

.bu-drawer-body {
  display: grid;
  gap: 16px;
  color: #f5f7fb;
}

.bu-template-list-view,
.bu-template-form-view {
  display: grid;
  gap: 16px;
}

.bu-drawer-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px;
  border: 1px solid rgba(125, 155, 190, 0.18);
  border-radius: 14px;
  color: #9aabc3;
  background: rgba(15, 26, 40, 0.58);
  font-size: 13px;
}

.bu-drawer-toolbar strong {
  display: block;
  margin-bottom: 5px;
  color: #f5f7fb;
  font-size: 16px;
  font-weight: 900;
}

.bu-drawer-section {
  padding: 18px;
  border: 1px solid rgba(125, 155, 190, 0.18);
  border-radius: 14px;
  background:
    linear-gradient(180deg, rgba(20, 37, 52, 0.88), rgba(16, 25, 39, 0.92)),
    #111827;
}

.bu-drawer-section h3 {
  margin: 0 0 14px;
  color: #16f2dc;
  font-size: 15px;
  font-weight: 900;
}

.bu-info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.bu-info-item {
  min-width: 0;
  padding: 12px;
  border-radius: 10px;
  background: rgba(8, 16, 28, 0.42);
}

.bu-info-item span {
  display: block;
  margin-bottom: 7px;
  color: #91a2ba;
  font-size: 12px;
  font-weight: 700;
}

.bu-info-item strong {
  display: block;
  overflow: hidden;
  color: #f5f7fb;
  font-size: 14px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bu-detail-remark {
  margin: 14px 0 0;
  color: #b7c5d8;
  font-size: 13px;
  line-height: 1.7;
}

.bu-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.bu-chip {
  max-width: 100%;
  padding: 7px 10px;
  overflow: hidden;
  border: 1px solid rgba(22, 242, 220, 0.26);
  border-radius: 999px;
  color: #16f2dc;
  background: rgba(22, 242, 220, 0.08);
  font-size: 12px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bu-detail-list {
  display: grid;
  gap: 10px;
}

.bu-detail-list-item {
  display: grid;
  gap: 6px;
  padding: 12px;
  border-radius: 10px;
  background: rgba(8, 16, 28, 0.42);
}

.bu-detail-list-item strong {
  color: #f5f7fb;
  font-size: 14px;
}

.bu-detail-list-item span,
.bu-detail-list-item em {
  color: #91a2ba;
  font-size: 12px;
  font-style: normal;
}

.bu-relation-list {
  display: grid;
  gap: 10px;
}

.bu-relation-row {
  display: grid;
  grid-template-columns: minmax(220px, 1.3fr) minmax(140px, 0.8fr) minmax(160px, 1fr) 64px;
  gap: 10px;
  align-items: center;
}

.bu-drawer-form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.bu-template-field-head,
.bu-template-field-row {
  display: grid;
  grid-template-columns: minmax(120px, 1fr) minmax(120px, 1fr) 112px minmax(110px, 0.9fr) minmax(130px, 1fr) 72px 62px;
  gap: 10px;
  align-items: center;
}

.bu-template-field-head {
  margin-bottom: 10px;
  color: #91a2ba;
  font-size: 12px;
  font-weight: 800;
}

.bu-template-field-list {
  display: grid;
  gap: 10px;
}

.bu-template-field-row {
  padding: 12px;
  border: 1px solid rgba(125, 155, 190, 0.18);
  border-radius: 12px;
  background: rgba(8, 16, 28, 0.34);
}

.bu-template-field-row.drawer {
  grid-template-columns: minmax(108px, 1.05fr) minmax(96px, 1fr) 90px minmax(88px, 0.9fr) minmax(100px, 1fr) 58px 44px;
  grid-template-areas: "label key type default placeholder required remove";
  gap: 10px;
  align-items: center;
  overflow: hidden;
}

.bu-template-field-row.drawer.has-options {
  grid-template-areas:
    "label key type default placeholder required remove"
    "options options options options options options options";
}

.bu-template-field-row.drawer :deep(.el-form-item) {
  margin-bottom: 0;
}

.bu-template-field-row.drawer :deep(.el-input),
.bu-template-field-row.drawer :deep(.el-select) {
  min-width: 0;
}

.bu-template-field-row.drawer .field-label {
  grid-area: label;
}

.bu-template-field-row.drawer .field-key {
  grid-area: key;
}

.bu-template-field-row.drawer .field-type {
  grid-area: type;
}

.bu-template-field-row.drawer .field-default {
  grid-area: default;
}

.bu-template-field-row.drawer .field-placeholder {
  grid-area: placeholder;
}

.field-required {
  grid-area: required;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-width: 0;
  color: #9aabc3;
  font-size: 12px;
  font-weight: 800;
  white-space: nowrap;
}

.field-remove {
  grid-area: remove;
  min-width: 0;
  padding: 0;
}

.bu-template-options {
  grid-area: options;
  width: 100%;
  min-width: 0;
}

.bu-drawer-footer {
  position: sticky;
  bottom: 0;
  z-index: 2;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 0 0;
  border-top: 1px solid rgba(125, 155, 190, 0.18);
  background: #111827;
}

.bu-save-error {
  flex: 1;
  min-width: 260px;
  padding: 9px 12px;
  border: 1px solid rgba(255, 100, 100, 0.35);
  border-radius: 10px;
  color: #ff7b7b;
  background: rgba(120, 24, 24, 0.22);
  font-size: 13px;
  font-weight: 800;
  text-align: left;
}

:deep(.bu-side-drawer) {
  background:
    linear-gradient(180deg, rgba(20, 37, 52, 0.98), rgba(16, 25, 39, 0.98)),
    #111827;
}

:deep(.bu-side-drawer .el-drawer__header) {
  margin: 0;
  padding: 22px 24px;
  border-bottom: 1px solid rgba(125, 155, 190, 0.18);
  color: #f5f7fb;
  font-weight: 900;
}

:deep(.bu-side-drawer .el-drawer__body) {
  padding: 18px;
  background: #111827;
}

@media (max-width: 1180px) {
  .bu-workspace {
    grid-template-columns: 1fr;
  }

  .bu-type-panel {
    min-height: 260px;
  }

  .bu-main {
    grid-template-rows: auto auto;
  }

  .bu-template-field-head {
    display: none;
  }

  .bu-template-field-row {
    grid-template-columns: 1fr;
  }

  .bu-template-field-row.drawer {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 96px 70px 48px;
    grid-template-areas:
      "label key type required remove"
      "default placeholder placeholder placeholder placeholder";
  }

  .bu-template-field-row.drawer.has-options {
    grid-template-areas:
      "label key type required remove"
      "default placeholder placeholder placeholder placeholder"
      "options options options options options";
  }

  .bu-actor-grid {
    grid-template-columns: 1fr;
  }
}
</style>
