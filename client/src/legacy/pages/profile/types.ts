export interface UserData {
  username: string;
  email: string;
  bio: string;
  level: number;
  vipLevel: number;
  userId: string;
  coins: number;
  gems: number;
  securityScore: number;
}

export interface ProfileContext {
  userData: UserData;
  setUserData: (data: UserData) => void;
  isEditing: boolean;
  setIsEditing: (editing: boolean) => void;
}
