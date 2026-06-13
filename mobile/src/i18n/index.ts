export type Lang = "am" | "om" | "ti" | "en";

const strings: Record<Lang, Record<string, string>> = {
  am: {
    welcome: "እንኳን ወደ መሶብ ዋሌት መጡ",
    balance: "ቀሪ ሂሳብ",
    cashIn: "ጥሬ ገንዘብ ያስገቡ",
    cashOut: "ጥሬ ገንዘብ ያውጡ",
  },
  om: {
    welcome: "Baga nagaan dhuftan Mesob Wallet",
    balance: "Hanga herregaa",
    cashIn: "Maallaqa galchi",
    cashOut: "Maallaqa baasi",
  },
  ti: {
    welcome: "ናብ መሶብ ወፍሪ እንቋዕ ብደሓን መጻእኩም",
    balance: "ዝተረፈ ሒሳብ",
    cashIn: "ናብ ሒሳብ ምእታው",
    cashOut: "ካብ ሒሳብ ምውጻእ",
  },
  en: {
    welcome: "Welcome to Mesob Wallet",
    balance: "Balance",
    cashIn: "Cash In",
    cashOut: "Cash Out",
  },
};

export function t(lang: Lang, key: string): string {
  return strings[lang][key] ?? strings.en[key] ?? key;
}
