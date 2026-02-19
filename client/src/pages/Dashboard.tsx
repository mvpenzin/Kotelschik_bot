import { useBotStatus, useBotLogs } from "@/hooks/use-bot";
import { Layout } from "@/components/Layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Activity,
  Clock,
  ShieldCheck,
  AlertCircle,
  RefreshCw,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";

export default function Dashboard() {
  const { data: status, isLoading: isLoadingStatus } = useBotStatus();
  const { data: logs, isLoading: isLoadingLogs } = useBotLogs();

  // Filter for error logs
  const errorLogs = logs?.filter((log) => log.level === "ERROR") || [];
  const recentLogs = logs?.slice(0, 5) || [];

  return (
    <Layout>
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight">Панель управления</h1>
        <p className="text-muted-foreground">
          Обзор состояния бота и его недавней активности
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="glass-panel border-l-4 border-l-emerald-500">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Состояние системы
            </CardTitle>
            <Activity className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            {isLoadingStatus ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <div className="flex items-center gap-2">
                <div
                  className={`w-3 h-3 rounded-full ${status?.status === "running" ? "bg-emerald-500 animate-pulse" : "bg-red-500"}`}
                />
                <div className="text-2xl font-bold capitalize">
                  {status?.status || "Unknown"}
                </div>
              </div>
            )}
            <p className="text-xs text-muted-foreground mt-2">
              Проверка:{" "}
              {status?.lastCheck
                ? formatDistanceToNow(new Date(status.lastCheck), {
                    addSuffix: true,
                  })
                : "..."}
            </p>
          </CardContent>
        </Card>

        <Card className="glass-panel border-l-4 border-l-indigo-500">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Доступность
            </CardTitle>
            <Clock className="h-4 w-4 text-indigo-500" />
          </CardHeader>
          <CardContent>
            {isLoadingStatus ? (
              <Skeleton className="h-8 w-32" />
            ) : (
              <div className="text-2xl font-bold">
                {status ? Math.floor(status.uptime / 3600) : 0}h{" "}
                {status ? Math.floor((status.uptime % 3600) / 60) : 0}m
              </div>
            )}
            <p className="text-xs text-muted-foreground mt-2">
              c момента последнего перезапуска
            </p>
          </CardContent>
        </Card>

        <Card className="glass-panel border-l-4 border-l-amber-500">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Диагностика
            </CardTitle>
            <ShieldCheck className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">Исправен</div>
            <p className="text-xs text-muted-foreground mt-2">
              Подключение к базе данных активно
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Live Logs Preview */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">Логи</h2>
            <Badge variant="outline" className="font-mono text-xs">
              online
            </Badge>
          </div>
          <Card className="bg-black/40 border-border/50 overflow-hidden">
            <div className="p-4 font-mono text-sm space-y-3 max-h-[400px] overflow-y-auto">
              {isLoadingLogs ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-4 w-full bg-white/5" />
                ))
              ) : recentLogs.length > 0 ? (
                recentLogs.map((log) => (
                  <div key={log.id} className="grid grid-cols-[auto_1fr] gap-3">
                    <span className="text-muted-foreground text-xs whitespace-nowrap pt-0.5">
                      {new Date(log.createdAt!).toLocaleTimeString()}
                    </span>
                    <div className="break-all">
                      <span
                        className={`
                        mr-2 font-bold text-xs px-1.5 py-0.5 rounded
                        ${log.level === "INFO" ? "bg-blue-500/10 text-blue-400" : ""}
                        ${log.level === "WARN" ? "bg-amber-500/10 text-amber-400" : ""}
                        ${log.level === "ERROR" ? "bg-red-500/10 text-red-400" : ""}
                      `}
                      >
                        {log.level}
                      </span>
                      <span className="text-gray-300">{log.message}</span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-muted-foreground text-center py-8">
                  Логи отсутствуют
                </div>
              )}
            </div>
          </Card>
        </div>

        {/* Action Required / Alerts */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">Системные оповещения</h2>
          </div>
          {errorLogs.length > 0 ? (
            <div className="space-y-3">
              {errorLogs.slice(0, 3).map((log) => (
                <div
                  key={log.id}
                  className="bg-red-950/20 border border-red-500/20 p-4 rounded-lg flex gap-3 items-start"
                >
                  <AlertCircle className="w-5 h-5 text-red-400 shrink-0 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-red-200 text-sm">
                      Обнаружена ошибка
                    </h3>
                    <p className="text-red-200/70 text-sm mt-1">
                      {log.message}
                    </p>
                    <p className="text-xs text-red-200/40 mt-2 font-mono">
                      {new Date(log.createdAt!).toLocaleString()}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-emerald-950/10 border border-emerald-500/20 p-8 rounded-lg flex flex-col items-center justify-center text-center">
              <div className="w-12 h-12 bg-emerald-500/20 rounded-full flex items-center justify-center mb-3">
                <ShieldCheck className="w-6 h-6 text-emerald-400" />
              </div>
              <h3 className="font-semibold text-emerald-200">
                Все системы в норме
              </h3>
              <p className="text-emerald-200/60 text-sm mt-1">
                Ошибок в логах не обнаружено
              </p>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}
