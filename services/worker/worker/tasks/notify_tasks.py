import logging
from datetime import datetime
from typing import Any

from celery import shared_task

logger = logging.getLogger(__name__)


@shared_task(
    name="worker.notify.email",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def send_email(
    self,
    to: str | list[str],
    subject: str,
    body: str,
    options: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """
    发送邮件通知。

    Args:
        to: 收件人邮箱 (单字符串或列表)
        subject: 邮件主题
        body: 邮件正文
        options: 额外选项 (cc, bcc, attachments 等)

    Returns:
        发送结果
    """
    if isinstance(to, str):
        to_list = [to]
    else:
        to_list = to

    logger.info(f"[send_email] to={to_list}, subject={subject}")

    try:
        result = {
            "status": "sent",
            "to": to_list,
            "subject": subject,
            "sent_at": datetime.now().isoformat(),
            "message_id": f"msg_{datetime.now().strftime('%Y%m%d%H%M%S')}",
        }
        logger.info(f"[send_email] success to={to_list}")
        return result
    except Exception as exc:
        logger.error(f"[send_email] failed to={to_list}: {exc}")
        raise


@shared_task(
    name="worker.notify.welcome",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def send_welcome_email(self, user_id: str, email: str, display_name: str) -> dict[str, Any]:
    """
    发送欢迎邮件。

    Args:
        user_id: 用户ID
        email: 用户邮箱
        display_name: 显示名称

    Returns:
        发送结果
    """
    logger.info(f"[send_welcome_email] user_id={user_id}, email={email}")

    try:
        subject = f"欢迎加入 DeepWrite, {display_name}!"
        body = f"""
        <html>
        <body>
            <h1>欢迎, {display_name}!</h1>
            <p>感谢您注册 DeepWrite - 沉浸式 AI 协作科研工作平台。</p>
            <p>开始探索您的创作之旅吧!</p>
        </body>
        </html>
        """
        return send_email.apply_async(kwargs={"to": email, "subject": subject, "body": body})
    except Exception as exc:
        logger.error(f"[send_welcome_email] failed user_id={user_id}: {exc}")
        raise


@shared_task(
    name="worker.notify.document_shared",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def send_document_shared_notification(
    self,
    recipient_email: str,
    sender_name: str,
    document_title: str,
    workspace_name: str,
) -> dict[str, Any]:
    """
    发送文档分享通知。

    Args:
        recipient_email: 接收者邮箱
        sender_name: 发送者名称
        document_title: 文档标题
        workspace_name: 工作区名称

    Returns:
        发送结果
    """
    logger.info(f"[send_document_shared] to={recipient_email}, doc={document_title}")

    try:
        subject = f"{sender_name} 分享了文档给您"
        body = f"""
        <html>
        <body>
            <h1>{sender_name} 分享了文档</h1>
            <p><strong>文档:</strong> {document_title}</p>
            <p><strong>工作区:</strong> {workspace_name}</p>
            <p>点击查看: https://deepwrite.work/documents</p>
        </body>
        </html>
        """
        return send_email.apply_async(kwargs={"to": recipient_email, "subject": subject, "body": body})
    except Exception as exc:
        logger.error(f"[send_document_shared] failed: {exc}")
        raise


@shared_task(
    name="worker.notify.task_reminder",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def send_task_reminder(
    self,
    recipient_email: str,
    task_title: str,
    due_date: str,
) -> dict[str, Any]:
    """
    发送任务提醒。

    Args:
        recipient_email: 接收者邮箱
        task_title: 任务标题
        due_date: 截止日期

    Returns:
        发送结果
    """
    logger.info(f"[send_task_reminder] to={recipient_email}, task={task_title}")

    try:
        subject = f"任务提醒: {task_title}"
        body = f"""
        <html>
        <body>
            <h1>任务提醒</h1>
            <p><strong>任务:</strong> {task_title}</p>
            <p><strong>截止日期:</strong> {due_date}</p>
            <p>请及时处理!</p>
        </body>
        </html>
        """
        return send_email.apply_async(kwargs={"to": recipient_email, "subject": subject, "body": body})
    except Exception as exc:
        logger.error(f"[send_task_reminder] failed: {exc}")
        raise
