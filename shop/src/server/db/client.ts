import pkg from "@prisma/client";
const { PrismaClient } = pkg as any;

declare global {
  // eslint-disable-next-line no-var
  var prisma: any | undefined;
}

export const prisma: any = global.prisma ?? new PrismaClient();

if (process.env.NODE_ENV !== "production") {
  global.prisma = prisma;
}
