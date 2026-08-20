# Bug reproduction

交付包写入器返回磁盘错误时，申请却已被标记为 downloaded。修复应在成功写包后再提交状态。
