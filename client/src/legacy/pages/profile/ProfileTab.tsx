import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Award, Edit3, Save, Shield } from "lucide-react";
import { cn } from "@/lib/utils";
import { useOutletContext } from "react-router-dom";
import type { ProfileContext } from "./types";

export default function ProfileTab() {
  const { userData, setUserData, isEditing, setIsEditing } =
    useOutletContext<ProfileContext>();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">基本资料</h3>
        <Button
          variant="outline"
          size="sm"
          onClick={() => (isEditing ? setIsEditing(false) : setIsEditing(true))}
          className="gap-1.5"
        >
          {isEditing ? (
            <>
              <Save className="w-4 h-4" />
              保存
            </>
          ) : (
            <>
              <Edit3 className="w-4 h-4" />
              编辑
            </>
          )}
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-2">
          <Label>昵称</Label>
          <Input
            value={userData.username}
            disabled={!isEditing}
            onChange={(e) =>
              setUserData({ ...userData, username: e.target.value })
            }
          />
        </div>
        <div className="space-y-2">
          <Label>邮箱</Label>
          <Input
            value={userData.email}
            disabled={!isEditing}
            onChange={(e) =>
              setUserData({ ...userData, email: e.target.value })
            }
          />
        </div>
        <div className="space-y-2 md:col-span-2">
          <Label>个人简介</Label>
          <Input
            value={userData.bio}
            disabled={!isEditing}
            onChange={(e) =>
              setUserData({ ...userData, bio: e.target.value })
            }
          />
        </div>
        <div className="space-y-2">
          <Label>用户 ID</Label>
          <Input value={userData.userId} disabled />
        </div>
        <div className="space-y-2">
          <Label>当前等级</Label>
          <div className="flex items-center gap-2 h-9 px-3 rounded-md border border-input bg-transparent text-sm">
            <Award className="w-4 h-4 text-primary" />
            Lv.{userData.level}
          </div>
        </div>
      </div>

      <Separator />

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">账号安全</h3>
        <div className="flex items-center gap-4">
          <div className="flex-1 max-w-xs">
            <div className="flex items-center justify-between mb-1.5">
              <span className="text-sm text-muted-foreground">安全评分</span>
              <span className="text-sm font-medium">
                {userData.securityScore}分
              </span>
            </div>
            <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
              <div
                className={cn(
                  "h-full rounded-full transition-all",
                  userData.securityScore >= 80
                    ? "bg-green-500"
                    : userData.securityScore >= 50
                      ? "bg-yellow-500"
                      : "bg-red-500",
                )}
                style={{ width: `${userData.securityScore}%` }}
              />
            </div>
          </div>
          <div className="flex items-center gap-1.5 text-sm text-green-500">
            <Shield className="w-4 h-4" />
            <span>安全</span>
          </div>
          <Button variant="link" size="sm" className="text-primary">
            修改密码
          </Button>
        </div>
      </div>

      <Separator />

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">我的特权</h3>
        <div className="flex items-center gap-3 flex-wrap">
          {[1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="w-10 h-10 rounded-lg bg-secondary flex items-center justify-center"
            >
              <Award className="w-5 h-5 text-primary/80" />
            </div>
          ))}
          <Button variant="ghost" size="sm" className="text-primary">
            更多 »
          </Button>
        </div>
      </div>
    </div>
  );
}
