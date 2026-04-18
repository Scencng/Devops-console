import { ElMessageBox } from 'element-plus'

export function showError(message,title) {
    ElMessageBox.alert(message, title, {
        type: 'warning',
        confirmButtonText: '确定'
    }).then(r => {
        window.location.href = '/login'
    })
}
