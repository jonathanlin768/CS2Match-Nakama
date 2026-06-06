"use client";

import { Clock, Check } from "lucide-react";

const tasks = [
  { name: "完成 2 场模拟对战", progress: "1/2", xp: 200, completed: false },
  { name: "使用冲锋枪击杀 20人", progress: "14/20", xp: 200, completed: false },
  { name: "赢得 1 场比赛", progress: "1/1", xp: null, completed: true },
];

export function DailyTasks() {
  return (
    <div className="flex-1 bg-card rounded-md border border-border p-3 lg:p-4 flex flex-col">
      <div className="flex items-center justify-between mb-3 lg:mb-4">
        <h3 className="font-semibold text-sm lg:text-base">今日任务</h3>
        <div className="flex items-center gap-1 text-[10px] lg:text-xs text-muted-foreground">
          <Clock className="w-3 lg:w-3.5 h-3 lg:h-3.5" />
          <span className="hidden sm:inline">12:45:23 后刷新</span>
          <span className="sm:hidden">12:45:23</span>
        </div>
      </div>

      <div className="space-y-2 lg:space-y-3 flex-1">
        {tasks.map((task, index) => (
          <div
            key={index}
            className="flex items-center justify-between p-2 lg:p-3 bg-secondary/50 rounded"
          >
            <div className="flex-1 min-w-0">
              <div className="text-xs lg:text-sm font-medium truncate">{task.name}</div>
              <div className="text-[10px] lg:text-xs text-muted-foreground mt-0.5">
                {task.progress}
              </div>
            </div>
            <div className="flex items-center gap-1.5 lg:gap-2 flex-shrink-0 ml-2">
              {task.xp && (
                <span className="px-1.5 lg:px-2 py-0.5 bg-primary/20 text-primary text-[10px] lg:text-xs rounded">
                  XP {task.xp}
                </span>
              )}
              {task.completed && (
                <div className="w-4 lg:w-5 h-4 lg:h-5 bg-green-500 rounded-full flex items-center justify-center">
                  <Check className="w-2.5 lg:w-3 h-2.5 lg:h-3 text-white" />
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      <button className="w-full mt-3 lg:mt-4 py-1.5 lg:py-2 text-xs lg:text-sm text-muted-foreground hover:text-foreground transition-colors border border-border rounded">
        查看全部任务
      </button>
    </div>
  );
}
