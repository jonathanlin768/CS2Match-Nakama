import type { Session, Socket, Channel, ChannelMessageAck, ChannelMessageList } from "@heroiclabs/nakama-js";
import client from "../nakama";

/**
 * 加入与指定好友的 1v1 DirectMessage 持久化频道。
 *
 * 调用 Nakama socket.joinChat，频道类型为 DirectMessage (type=2)，
 * persistence=true 使消息持久化到 PostgreSQL，hidden=false 使频道可见。
 * Nakama 内部使用排序后的两个用户 ID 生成确定性 DM channel_id，
 * 双方 join 同一个目标用户 ID 会进入相同频道。
 *
 * @param socket - 已连接的 Nakama WebSocket 实例
 * @param targetUserId - 目标好友的 Nakama 用户 ID
 * @returns Channel 对象（含 id、presences、self）
 */
export async function joinDMChannel(
  socket: Socket,
  targetUserId: string,
): Promise<Channel> {
  // type=2 DirectMessage, persistence=true, hidden=false
  return socket.joinChat(targetUserId, 2, true, false);
}

/**
 * 向指定频道发送文本消息。
 *
 * 调用 socket.writeChatMessage 发送消息到指定频道。
 * Nakama 会将消息广播给频道内所有成员（包括发送者自己）通过 onchannelmessage 推送。
 * 消息内容格式为 { text: string }，便于未来扩展富媒体。
 *
 * @param socket - 已连接的 Nakama WebSocket 实例
 * @param channelId - 目标频道 ID
 * @param content - 消息文本内容
 * @returns ChannelMessageAck（含 message_id、channel_id、create_time 等）
 */
export async function writeChatMessage(
  socket: Socket,
  channelId: string,
  content: string,
): Promise<ChannelMessageAck> {
  return socket.writeChatMessage(channelId, { text: content });
}

/**
 * 获取指定频道的消息历史（REST API）。
 *
 * 调用 client.listChannelMessages 获取已持久化的历史消息。
 * 支持分页：通过 cursor 和 forward 参数向前翻页加载更早的消息。
 *
 * @param session - 当前有效的 Nakama Session
 * @param channelId - 频道 ID
 * @param limit - 单页消息数（默认 100）
 * @param forward - true=向前翻页（加载更早消息），false/undefined=向后
 * @param cursor - 分页游标
 * @returns ChannelMessageList（含 messages 数组和 next_cursor）
 */
export async function listChannelMessages(
  session: Session,
  channelId: string,
  limit: number = 100,
  forward?: boolean,
  cursor?: string,
): Promise<ChannelMessageList> {
  return client.listChannelMessages(session, channelId, limit, forward, cursor);
}

/**
 * 退出指定的 DM 频道。
 *
 * 调用 socket.leaveChat 退出频道，之后不再收到该频道的消息推送。
 * 仅在好友关系解除时调用（删除好友后退出其 DM 频道）。
 *
 * @param socket - 已连接的 Nakama WebSocket 实例
 * @param channelId - 要退出的频道 ID
 */
export async function leaveDMChannel(
  socket: Socket,
  channelId: string,
): Promise<void> {
  return socket.leaveChat(channelId);
}
