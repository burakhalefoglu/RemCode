# Araştırma Brief'i #2 — Desteklenen Araçlar Listesine Nasıl Giriliyor?

> **Neden bu araştırma:** Üç sağlayıcının da başvuru kanalı ya yok ya ulaşılamıyor.
> Başvurmadan önce **başvurulacak bir şey olup olmadığını** doğrulamak gerekiyor.
>
> **Ana hipotez:** Liste **betimleyici** (sağlayıcı popüler araçları gözlemleyip
> dokümante ediyor), **izin verici** değil (araçlar başvurup ekleniyor değil).
>
> **En güçlü ilk kanıt:** Z.AI'nin coding plan listesinde **SillyTavern** var —
> bir roleplay sohbet arayüzü, coding aracı bile değil. Kimsenin başvurduğu
> düşünülemez.

**Tarih:** ……………  **Araştıran:** ……………

---

## R1 — Liste betimleyici mi, izin verici mi?

Listedeki araçlardan birkaçını seç (**Cline, Roo Code, OpenCode, Kilo Code,
Crush, Goose** — hepsi açık kaynak, geçmişleri okunabilir) ve şuna bak:

| # | Soru | Bulgu |
| :-- | :--- | :--- |
| R1.1 | Bu araçlardan **herhangi biri** Z.AI/Qwen/Kimi ile ortaklık duyurdu mu? (blog, changelog, README, X) | |
| R1.2 | Yoksa sağlayıcı mı tek taraflı "X aracını destekliyoruz" dedi? | |
| R1.3 | Bir aracın "listeye eklenmek için başvurduk" dediği **tek bir** kamuya açık kayıt var mı? | |
| R1.4 | Listeye yakın zamanda yeni araç eklendi mi? Hangisi, ne zaman, nasıl? (docs git history / web archive) | |
| R1.5 | **SillyTavern anomalisi:** SillyTavern tarafında Z.AI ile herhangi bir temas izi var mı? | |

**Karar kuralı:** R1.3 boşsa ve R1.2 doluysa → liste betimleyici → **başvuru
diye bir şey yok, aramayı bırak.**

---

## R2 — Teknik kapı var mı? (EN KRİTİK BÖLÜM)

Asıl soru bu: sağlayıcı istemciyi **teknik olarak** mı tanıyor, yoksa sadece
dokümanda mı yazıyor?

Listedeki açık kaynak araçların **kaynak kodunu oku.** Z.AI / Qwen coding /
Kimi coding endpoint'ine giden isteklerde:

| # | Soru | Bulgu |
| :-- | :--- | :--- |
| R2.1 | Özel bir **User-Agent** gönderiyorlar mı? Ne? | |
| R2.2 | `client_id`, `X-Client`, `X-Source` gibi bir **kimlik header'ı** var mı? | |
| R2.3 | Sağlayıcıya özel **herhangi bir** ek başlık/parametre var mı, yoksa düz OpenAI çağrısı mı? | |
| R2.4 | Kod, endpoint'e göre davranış değiştiriyor mu, yoksa "custom base URL" alanına ne yazılırsa oraya mı gidiyor? | |
| R2.5 | Aracın dokümanı kullanıcıya **"Coding Plan anahtarını buraya yapıştır"** diyor mu? | |

**Nerede bakılacak:**
- Cline → `src/api/providers/` altında OpenAI-compatible sağlayıcı
- Roo Code / Kilo Code → Cline fork'ları, aynı yapı
- OpenCode, Crush, Goose → provider/auth katmanı
- Qwen Code → `packages/core` içinde content generator

**Karar kuralı:**
- R2.1–R2.3 boşsa → **teknik kapı yok.** Endpoint kimliği kontrol etmiyor;
  kural dokümanda yaşıyor, kodda değil.
- Doluysa → o başlığı bilmeden zaten çalışmaz; ve taklit etmek **yasak**
  (Kimi'nin "client identity spoofing" maddesi). O zaman gerçekten kapı var.

---

## R3 — Yaptırım gerçekte uygulanıyor mu?

Kural yazılı olmak zorunda değil, **uygulanıyor** olmalı ki risk gerçek olsun.

| # | Soru | Bulgu |
| :-- | :--- | :--- |
| R3.1 | "Listede olmayan araçla coding plan kullandım, hesabım askıya alındı" diyen **gerçek** rapor var mı? (GitHub issue, Reddit, forum) | |
| R3.2 | Z.AI'nin 1302/1303 hata kodlarını alan kullanıcılar ne yapıyordu? Araç mı, kullanım deseni mi tetikledi? | |
| R3.3 | Yaptırım **araç kimliğine** mi bakıyor, **kullanım desenine** mi (hacim, coding-dışı içerik, eşzamanlılık)? | |
| R3.4 | Askıya alınıp itiraz edip geri dönen var mı? Süreç ne kadar sürüyor? | |

**Neden önemli:** Yaptırım desen bazlıysa, dürüst bir coding aracı olarak
davranmak yeterli olabilir. Kimlik bazlıysa liste gerçekten kapı demektir.

---

## R4 — Araçların kendi konumu

| # | Soru | Bulgu |
| :-- | :--- | :--- |
| R4.1 | Bu araçlar dokümanlarında sağlayıcı aboneliği kullanımına dair **uyarı/feragat** koyuyor mu? | |
| R4.2 | "ToS'a uymak kullanıcının sorumluluğu" tarzı bir ifade var mı? | |
| R4.3 | Yoksa hiç değinmiyorlar mı? | |

**Neden önemli:** Hiçbiri değinmiyorsa, ekosistemin fiilî normu şu demektir:
**araç genel BYOK yeteneği sunar, hangi anahtarın yapıştırıldığı kullanıcının
sorunudur.** Bizim pozisyonumuz da o olur.

---

## R5 — Karşı örnek avı (kendi hipotezimi kırmaya çalış)

Hipotezi doğrulayan kanıt aramak kolay. Asıl iş onu **çürütecek** kanıt aramak:

| # | Soru | Bulgu |
| :-- | :--- | :--- |
| R5.1 | Bir araç geliştiricisinin "sağlayıcı bizi listeye ekledi çünkü konuştuk" dediği bir örnek var mı? | |
| R5.2 | Sağlayıcıların bir "partner program" / "integration program" sayfası var mı? (docs dışında, kurumsal sitede) | |
| R5.3 | Listeye eklenen bir aracın eklenme tarihi ile o araçtaki bir sağlayıcı-özel değişikliğin tarihi çakışıyor mu? | |

---

## Karar tablosu — araştırma bittiğinde doldur

| Senaryo | R1.3 | R2.1–2.3 | R3.3 | Sonuç |
| :--- | :--- | :--- | :--- | :--- |
| **A — Kapı yok** | boş | boş | desen | Başvuru aramayı bırak. Genel BYOK aracı ol, PAYG'ı dokümante et, abonelik kullanımını **pazarlama**. Kullanıcı ne yapıştırırsa kendi ToS ilişkisi. |
| **B — Yumuşak kapı** | boş | boş | kimlik | Teknik engel yok ama yaptırım kimliğe bakıyor. Riskli — kullanıcıyı uyar, varsayılan PAYG. |
| **C — Sert kapı** | dolu | dolu | kimlik | Gerçek başvuru süreci var. Kanalı R5.2'den bul, başvur. |

---

## Bu araştırmanın cevaplamadığı şey (bilerek)

**"Yapabilir miyiz" ile "yapmalı mıyız" ayrı sorular.** Bu brief birincisini
ölçüyor. İkincisine A senaryosunda bile dikkat: sağlayıcının yazılı kuralını
teknik olarak aşabiliyor olmak, aşmayı doğru yapmaz. Ama **kullanıcının kendi
anahtarını kendi makinesinde kullanması** ile **bizim onun kotasını tüketmemiz**
farklı şeyler — ve tüm ekosistem bu ayrımın üzerinde duruyor.

Sonuç ne çıkarsa çıksın, ürünün varsayılanı **PAYG API anahtarı** kalır.
