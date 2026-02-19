import { Link, useLocation } from "wouter";
import { LayoutDashboard, Users, Terminal, Activity } from "lucide-react";

export function Layout({ children }: { children: React.ReactNode }) {
  const [location] = useLocation();

  const navItems = [
    { href: "/", label: "Панель управления", icon: LayoutDashboard },
    { href: "/contacts", label: "Контакты", icon: Users },
    { href: "/details", label: "Реквизиты", icon: Users },
    { href: "/prices", label: "Тарифы", icon: Users },
    { href: "/debts", label: "Задолженности", icon: Users },
    { href: "/logs", label: "Системные логи", icon: Terminal },
  ];

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside className="w-64 border-r border-border bg-card/30 hidden md:flex flex-col">
        <div className="p-6 border-b border-border/50">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-indigo-500/20 flex items-center justify-center text-indigo-400">
              <Activity className="w-5 h-5" />
            </div>
            <h1 className="font-bold text-lg tracking-tight">
              СНТ "Котельщик"
            </h1>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => {
            const isActive = location === item.href;
            return (
              <Link key={item.href} href={item.href}>
                <div
                  className={`
                    flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 cursor-pointer
                    ${
                      isActive
                        ? "bg-primary text-primary-foreground shadow-lg shadow-black/10"
                        : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                    }
                  `}
                >
                  <item.icon
                    className={`w-4 h-4 ${isActive ? "text-primary-foreground" : "text-muted-foreground"}`}
                  />
                  {item.label}
                </div>
              </Link>
            );
          })}
        </nav>

        <div className="p-4 border-t border-border/50">
          <div className="bg-muted/30 rounded-xl p-4">
            <p className="text-xs text-muted-foreground font-mono">
              v0.0.1-beta
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              Статус: В работе
            </p>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Mobile Header (visible only on small screens) */}
        <header className="md:hidden h-16 border-b border-border flex items-center px-4 justify-between bg-card">
          <div className="font-bold">СНТ "Котельщик"</div>
        </header>

        <div className="flex-1 overflow-auto p-4 md:p-8 lg:p-12">
          <div className="max-w-7xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {children}
          </div>
        </div>
      </main>
    </div>
  );
}
