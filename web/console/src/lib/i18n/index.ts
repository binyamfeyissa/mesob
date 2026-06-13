export type Lang = "am" | "om" | "ti" | "en";

const strings: Record<Lang, Record<string, string>> = {
  am: {
    dashboard: "ዋና ገጽ",
    riskFraud: "ስጋት እና ማጭበርበር",
    settlements: "ማስፈፀሚያ",
    kycReview: "KYC ግምገማ",
    auditLog: "ኦዲት ምዝግብ ማስታወሻ",
    config: "ውቅር",
    live: "ቀጥታ",
    disputes: "አለመግባባቶች",
    reconciliation: "ሂሳብ ማስታረቅ",
    approve: "ፍቀድ",
    reject: "ውድቅ",
    refund: "መልስ",
    deny: "ክልክል",
    clear: "አጽዳ",
    confirm: "አረጋግጥ",
    escalateSar: "SAR አሳድግ",
    amountETB: "መጠን (ብር)",
    secondAuthoriser: "ሁለተኛ ፈቃድ ሰጪ",
    note: "ማስታወሻ",
    submit: "ላክ",
    loading: "በመጫን ላይ...",
    noData: "ምንም ውሂብ የለም",
  },
  om: {
    dashboard: "Fuula Ijoo",
    riskFraud: "Balaa fi Doorsisaa",
    settlements: "Qindeessuu",
    kycReview: "Sakatta'a KYC",
    auditLog: "Galmee Qorannoo",
    config: "Qindaa'ina",
    live: "Kallattii",
    disputes: "Mormii",
    reconciliation: "Waldorgommii",
    approve: "Hayyami",
    reject: "Didu",
    refund: "Deebisi",
    deny: "Dhowwi",
    clear: "Haqi",
    confirm: "Mirkaneessi",
    escalateSar: "SAR Dabali",
    amountETB: "Gatii (ETB)",
    secondAuthoriser: "Kan lammaffaa hayyamee",
    note: "Yaadannoo",
    submit: "Ergi",
    loading: "Fe'aa jira...",
    noData: "Odeeffannoon hin jiru",
  },
  ti: {
    dashboard: "ዋና ገጽ",
    riskFraud: "ሓደጋን ምትላልን",
    settlements: "ምፍጻም",
    kycReview: "ምግምጋም KYC",
    auditLog: "መዝገብ ኦዲት",
    config: "ቅርጺ",
    live: "ቀጥታ",
    disputes: "ምፍልላያት",
    reconciliation: "ምምዝናን",
    approve: "ፍቐድ",
    reject: "ንጸስ",
    refund: "መልስ",
    deny: "ኣይፋል",
    clear: "ጽረግ",
    confirm: "ኣረጋግጽ",
    escalateSar: "SAR ዕበ",
    amountETB: "መጠን (ብር)",
    secondAuthoriser: "ካልኣይ ፍቓድ ሂቦ",
    note: "ዝኽሪ",
    submit: "ስደድ",
    loading: "ይጽዓን ኣሎ...",
    noData: "ዝኾነ ሓበሬታ የለን",
  },
  en: {
    dashboard: "Dashboard",
    riskFraud: "Risk & Fraud",
    settlements: "Settlements",
    kycReview: "KYC Review",
    auditLog: "Audit Log",
    config: "Config",
    live: "Live",
    disputes: "Disputes",
    reconciliation: "Reconciliation",
    approve: "Approve",
    reject: "Reject",
    refund: "Refund",
    deny: "Deny",
    clear: "Clear",
    confirm: "Confirm",
    escalateSar: "Escalate SAR",
    amountETB: "Amount (ETB)",
    secondAuthoriser: "Second Authoriser ID",
    note: "Note",
    submit: "Submit",
    loading: "Loading...",
    noData: "No data",
  },
};

let _lang: Lang = "en";

export function setLang(lang: Lang): void {
  _lang = lang;
}

export function getLang(): Lang {
  return _lang;
}

export function t(key: string): string {
  return strings[_lang][key] ?? strings.en[key] ?? key;
}
