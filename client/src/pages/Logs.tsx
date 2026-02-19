import { useState } from "react";
import { Layout } from "@/components/Layout";
import { useBotLogs } from "@/hooks/use-bot";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Search, Filter, Terminal } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { format } from "date-fns";

export default function Logs() {
  const { data: logs, isLoading } = useBotLogs();
  const [searchTerm, setSearchTerm] = useState("");
  const [levelFilter, setLevelFilter] = useState<string>("ALL");

  const filteredLogs =
    logs?.filter((log) => {
      const matchesSearch =
        log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.details?.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesLevel = levelFilter === "ALL" || log.level === levelFilter;
      return matchesSearch && matchesLevel;
    }) || [];

  return (
    <Layout>
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight">Системные логи</h1>
        <p className="text-muted-foreground">Подробные логи активности.</p>
      </div>

      <div className="glass-panel rounded-xl border border-border/50 flex flex-col h-[600px]">
        {/* Controls */}
        <div className="p-4 border-b border-border/50 bg-muted/20 flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="Поиск..."
              className="pl-9 bg-background/50 border-border/50"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div className="w-[180px]">
            <Select value={levelFilter} onValueChange={setLevelFilter}>
              <SelectTrigger className="bg-background/50 border-border/50">
                <Filter className="w-4 h-4 mr-2 text-muted-foreground" />
                <SelectValue placeholder="Уровень" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">Все</SelectItem>
                <SelectItem value="INFO">Информация</SelectItem>
                <SelectItem value="WARN">Предупреждения</SelectItem>
                <SelectItem value="ERROR">Ошибки</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Logs Terminal */}
        <div className="flex-1 bg-black/40 overflow-hidden font-mono text-sm">
          <ScrollArea className="h-full">
            <div className="p-4 space-y-1">
              {isLoading ? (
                <div className="flex items-center justify-center h-48 text-muted-foreground">
                  <Terminal className="w-6 h-6 mr-2 animate-pulse" />
                  Получение логов...
                </div>
              ) : filteredLogs.length === 0 ? (
                <div className="flex items-center justify-center h-48 text-muted-foreground">
                  Не найдены логи по вашим критериям.
                </div>
              ) : (
                filteredLogs.map((log) => (
                  <div
                    key={log.id}
                    className="group hover:bg-white/5 p-1 rounded -mx-1 px-2 flex gap-4 items-start transition-colors"
                  >
                    <span className="text-muted-foreground/50 w-[150px] shrink-0 select-none">
                      {log.createdAt
                        ? format(new Date(log.createdAt), "yyyy-MM-dd HH:mm:ss")
                        : "-"}
                    </span>
                    <Badge
                      variant="outline"
                      className={`
                        w-16 justify-center shrink-0 text-[10px] font-bold border-0
                        ${log.level === "INFO" ? "bg-blue-500/20 text-blue-400" : ""}
                        ${log.level === "WARN" ? "bg-amber-500/20 text-amber-400" : ""}
                        ${log.level === "ERROR" ? "bg-red-500/20 text-red-400" : ""}
                      `}
                    >
                      {log.level}
                    </Badge>
                    <div className="flex-1 break-all">
                      <span className="text-gray-300">{log.message}</span>
                      {log.details && (
                        <div className="mt-1 text-xs text-muted-foreground bg-black/20 p-2 rounded block whitespace-pre-wrap">
                          {log.details}
                        </div>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          </ScrollArea>
        </div>

        <div className="p-2 border-t border-border/50 bg-muted/20 text-xs text-muted-foreground flex justify-between px-4">
          <span>{filteredLogs.length} entries</span>
          <span>Filtered from {logs?.length || 0} total</span>
        </div>
      </div>
    </Layout>
  );
}
