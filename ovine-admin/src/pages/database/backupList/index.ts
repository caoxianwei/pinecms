export const schema = {
  type: 'page',
  title: "备份列表",
  body: {
    type: 'crud',
    api: '$preset.apis.list',
    draggable: true,
    bulkActions: [
      {
        label: '优化',
        size: 'xs',
        actionType: "ajax",
        confirmText: "确定要优化选中数据表吗? 这可能需要几分钟!",
        api: '$preset.apis.backup'
      },
      {
        label: '修复',
        size: 'xs',
        actionType: "ajax",
        confirmText: "确定要修复选中数据表吗? 这可能需要几分钟!",
        api: '$preset.apis.backup'
      },
    ],
    headerToolbar: [
      {
        type: 'columns-toggler',
        align: 'left',
      },
      '$preset.actions.backup',
      'bulkActions',
    ],
    columns: [
      {
        name: 'name',
        label: '备份文件名',
        type: 'text',
      },
    ],
  },
  preset: {
    actions: {
      backup: {
        limits: 'del',
        type: 'action',
        label: '备份数据库',
        size: 'xs',
        align: 'right',
        actionType: 'ajax',
        confirmText: "确定要备份数据库吗? 这可能需要几分钟!",
        api: "$preset.apis.backup"
      },
    }
  }
}
