import { useState } from "react";
import { Layout } from "@/components/Layout";
import {
  useContacts,
  useCreateContact,
  useDeleteContact,
} from "@/hooks/use-bot";
import { insertSntContactSchema, type InsertSntContact } from "@shared/schema";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Plus, Trash2, Search, Phone, Mail, User } from "lucide-react";
import { format } from "date-fns";

export default function Contacts() {
  const { data: contacts, isLoading } = useContacts();
  const [searchTerm, setSearchTerm] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const filteredContacts =
    contacts?.filter(
      (contact) =>
        contact.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        contact.description?.toLowerCase().includes(searchTerm.toLowerCase()),
    ) || [];

  return (
    <Layout>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Контакты</h1>
          <p className="text-muted-foreground">Управление контактами СНТ</p>
        </div>
        <CreateContactDialog
          open={isCreateOpen}
          onOpenChange={setIsCreateOpen}
        />
      </div>

      <div className="glass-panel rounded-xl border border-border/50 overflow-hidden">
        {/* Toolbar */}
        <div className="p-4 border-b border-border/50 bg-muted/20 flex gap-4">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="Поиск..."
              className="pl-9 bg-background/50 border-border/50 focus-visible:ring-1"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent border-border/50">
                <TableHead className="w-[10px]">#</TableHead>
                <TableHead className="w-[200px]">Тип</TableHead>
                <TableHead className="min-w-[200px]">Значение</TableHead>
                <TableHead>Дополнительно</TableHead>
                <TableHead>Комментарий</TableHead>
                <TableHead className="w-[100px] text-right">Действия</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center">
                    Загрузка контактов...
                  </TableCell>
                </TableRow>
              ) : filteredContacts.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="h-48 text-center text-muted-foreground"
                  >
                    Контакты не найдены.
                  </TableCell>
                </TableRow>
              ) : (
                filteredContacts.map((contact) => (
                  <TableRow
                    key={contact.prior}
                    className="hover:bg-muted/30 border-border/50"
                  >
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                          <User className="w-4 h-4" />
                        </div>
                        {contact.type}
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {contact.value}
                    </TableCell>
                    <TableCell>
                      {contact.phone && (
                        <div className="flex items-center gap-2 text-sm">
                          <Phone className="w-3 h-3 text-muted-foreground" />
                          {contact.phone}
                        </div>
                      )}
                    </TableCell>
                    <TableCell>
                      {contact.email && (
                        <div className="flex items-center gap-2 text-sm">
                          <Mail className="w-3 h-3 text-muted-foreground" />
                          {contact.email}
                        </div>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <DeleteContactButton
                        id={contact.id}
                        name={contact.name}
                      />
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </Layout>
  );
}

function CreateContactDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { mutate, isPending } = useCreateContact();

  const form = useForm<InsertSntContact>({
    resolver: zodResolver(insertSntContactSchema),
    defaultValues: {
      name: "",
      description: "",
      phone: "",
      email: "",
    },
  });

  const onSubmit = (data: InsertSntContact) => {
    mutate(data, {
      onSuccess: () => {
        onOpenChange(false);
        form.reset();
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button className="bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/20">
          <Plus className="w-4 h-4 mr-2" />
          Добавить
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px] bg-card border-border">
        <DialogHeader>
          <DialogTitle>Добавить контакт</DialogTitle>
          <DialogDescription>Создание нового контакта СНТ</DialogDescription>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 pt-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              {...form.register("name")}
              placeholder="e.g. Security Post"
            />
            {form.formState.errors.name && (
              <p className="text-xs text-red-500">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              {...form.register("description")}
              placeholder="Role or location"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="phone">Phone</Label>
              <Input
                id="phone"
                {...form.register("phone")}
                placeholder="+7..."
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                {...form.register("email")}
                placeholder="email@example.com"
              />
            </div>
          </div>
          <DialogFooter className="pt-4">
            <Button type="submit" disabled={isPending}>
              {isPending ? "Creating..." : "Create Contact"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function DeleteContactButton({ id, name }: { id: number; name: string }) {
  const { mutate, isPending } = useDeleteContact();

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="hover:bg-red-500/10 hover:text-red-500"
        >
          <Trash2 className="w-4 h-4" />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent className="bg-card border-border">
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Contact?</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete <strong>{name}</strong>? This action
            cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => mutate(id)}
            className="bg-red-500 hover:bg-red-600 text-white"
            disabled={isPending}
          >
            {isPending ? "Deleting..." : "Delete"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
