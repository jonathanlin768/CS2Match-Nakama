import type { Session, Friends } from "@heroiclabs/nakama-js";
import client from "../nakama";

/**
 * 列出当前用户的所有好友。
 *
 * 调用 Nakama listFriends API，默认 limit=100 确保一次加载全部好友（实验项目 <100 人在线）。
 *
 * @param session - 当前有效的 Nakama Session
 * @param limit - 单次返回的最大好友数（默认 100）
 * @returns 好友列表对象，含 friends 数组和可选的 cursor 分页游标
 */
export async function listFriends(
  session: Session,
  limit: number = 100,
): Promise<Friends> {
  return client.listFriends(session, undefined, limit);
}

/**
 * 通过用户名添加好友。
 *
 * 调用 Nakama addFriends API，发送好友请求。
 * 成功后对方将收到一条 state=INVITE_RECEIVED 的好友记录。
 *
 * @param session - 当前有效的 Nakama Session
 * @param username - 要添加的用户名（精确匹配）
 * @returns 操作成功返回 true
 */
export async function addFriendsByUsername(
  session: Session,
  username: string,
): Promise<boolean> {
  return client.addFriends(session, undefined, [username]);
}

/**
 * 删除好友关系。
 *
 * 调用 Nakama deleteFriends API，删除指定用户的好友关系。
 * 可用于删除好友、取消已发送请求、拒绝收到的请求。
 *
 * @param session - 当前有效的 Nakama Session
 * @param userId - 要删除的用户 ID
 * @returns 操作成功返回 true
 */
export async function deleteFriend(
  session: Session,
  userId: string,
): Promise<boolean> {
  return client.deleteFriends(session, [userId]);
}
