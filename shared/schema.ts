import { pgTable, text, serial, timestamp, boolean } from "drizzle-orm/pg-core";
import { createInsertSchema } from "drizzle-zod";
import { comment } from "postcss";
import { number, z } from "zod";

// --- Users Table (from Telegram) ---
export const sntUsers = pgTable("snt_users", {
  id: serial("id").primaryKey(),
  telegramId: text("telegram_id").notNull().unique(), // Store as text to avoid overflow with large IDs
  username: text("username"),
  firstName: text("first_name"),
  lastName: text("last_name"),
  createdAt: timestamp("created_at").defaultNow(),
});

// --- Contacts Table ---
export const sntContacts = pgTable("snt_contacts", {
  prior: serial("prior").primaryKey(),
  type: text("type").notNull(),
  value: text("value"),
  adds: text("adds"),
  comment: text("comment"),
  created: timestamp("created").defaultNow(),
  modified: timestamp("modified").defaultNow(),
});

// --- Bot Logs (for frontend display) ---
export const botLogs = pgTable("bot_logs", {
  id: serial("id").primaryKey(),
  level: text("level").notNull(), // "INFO", "WARN", "ERROR"
  message: text("message").notNull(),
  details: text("details"),
  createdAt: timestamp("created_at").defaultNow(),
});

// --- Schemas ---
export const insertSntUserSchema = createInsertSchema(sntUsers).omit({
  id: true,
  createdAt: true,
});
export const insertSntContactSchema = createInsertSchema(sntContacts).omit({
  prior: true,
  created: true,
});
export const insertBotLogSchema = createInsertSchema(botLogs).omit({
  id: true,
  createdAt: true,
});

// --- Types ---
export type SntUser = typeof sntUsers.$inferSelect;
export type InsertSntUser = z.infer<typeof insertSntUserSchema>;

export type SntContact = typeof sntContacts.$inferSelect;
export type InsertSntContact = z.infer<typeof insertSntContactSchema>;

export type BotLog = typeof botLogs.$inferSelect;
export type InsertBotLog = z.infer<typeof insertBotLogSchema>;
