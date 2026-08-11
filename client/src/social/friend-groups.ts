export function groupFriends<T extends { state?: number }>(friends: T[]) {
  return {
    accepted: friends.filter((friend) => friend.state === 0),
    outgoing: friends.filter((friend) => friend.state === 1),
    incoming: friends.filter((friend) => friend.state === 2),
  }
}
