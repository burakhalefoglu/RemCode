# Araştırma Brief'i — Sağlayıcı Erişim Modeli

> **Amaç:** [SPEC-provider-contract-06](../specs/provider-contract.md) ve
> [-07](../specs/provider-contract.md)'yi cevaplamak. Bu bulgular P1'in kimlik
> doğrulama tasarımını belirler.
>
> **Bu faz masa başı araştırması** — kod gerekmez. Ölçüm gerektiren sorular
> (tool şeması, streaming delta) ayrı; onlar anahtar + kod istiyor.
>
> **Doldur, geri getir.** Emin olmadığın yere `?` yaz — tahmin yazma. Bir
> `?`, yanlış bir "evet"ten çok daha ucuz.

**Tarih:** ……………  **Araştıran:** ……………

---

## 0. Önce bunu bul — en büyük kısayol

**Sağlayıcının resmî açık kaynak CLI/coding aracı var mı?**

| Sağlayıcı | Var mı? | Repo URL | Not |
| :--- | :--- | :--- | :--- |
| Z.AI | | | |
| Qwen | | | |
| Kimi | | | |

Varsa, **auth akışının en iyi dokümantasyonu o reponun kaynak kodudur.** Login
kodunu okumak, dokümantasyondan çok daha kesin cevap verir. Bulursan aşağıdaki
A bölümünün yarısı oradan çıkar.

---

## A. Erişim modeli — mimariyi bu belirliyor

Her sağlayıcı için A / B / C'den birine düşecek:

- **A tipi** — Abonelik sana **API anahtarı** veriyor → *bugünkü mimari zaten
  çalışıyor, sadece dokümantasyon işi*
- **B tipi** — Abonelik sadece **login/OAuth** ile → *gerçek mühendislik:
  device flow, token refresh, başsız re-auth*
- **C tipi** — Üçüncü taraf istemciye abonelik erişimi **yok** → *o sağlayıcı
  sadece pay-as-you-go API anahtarıyla çalışır*

### Doldurulacak tablo

| # | Soru | Z.AI | Qwen | Kimi |
| :-- | :--- | :--- | :--- | :--- |
| A1 | Sabit ücretli abonelik / coding plan var mı? | | | |
| A2 | Adı ve fiyat basamakları? | | | |
| A3 | **Abonelik API anahtarı veriyor mu?** (evet → A tipi) | | | |
| A4 | Sadece login ile mi erişiliyor? (evet → B tipi) | | | |
| A5 | B ise akış tipi: device code / tarayıcı yönlendirme / session token yapıştırma? | | | |
| A6 | B ise token ömrü ne kadar? | | | |
| A7 | B ise refresh mekanizması var mı, nasıl? | | | |
| A8 | Abonelik endpoint'i pay-as-you-go'dan farklı mı? | | | |
| A9 | Abonelikte hangi modeller dahil, hangileri değil? | | | |
| A10 | Kota/limit nasıl ifade ediliyor? (istek/gün, token/ay, eşzamanlılık) | | | |
| | **→ TİP (A/B/C)** | | | |

---

## B. Kullanım şartları — özelliği öldürebilecek olan

> Bu bölüm teknik fizibiliteden **daha önemli.** Yanlış tarafta olmak hukuki
> sorun değil, **kullanıcı hesabı banı** üretir — ve suçu bize keserler.
> Teknik olarak mümkün ama şartlarca yasaksa, o yol açılmaz.

| # | Soru | Z.AI | Qwen | Kimi |
| :-- | :--- | :--- | :--- | :--- |
| B1 | Şartlar, abonelik kimlik bilgisinin **üçüncü taraf istemcide** kullanımına izin veriyor mu? | | | |
| B2 | **Destekleniyor / belirsiz / yasak** — hangisi? | | | |
| B3 | İlgili maddenin **tam metni** (kopyala) | | | |
| B4 | **URL + eriştiğin tarih** | | | |
| B5 | Üçüncü taraf coding agent'lara dair **açık bir ifade** var mı? (bazıları bunu aktif pazarlıyor) | | | |
| B6 | Otomatik/arka plan kullanımı kısıtlayan fair-use maddesi var mı? | | | |
| B7 | Hesap paylaşımı / kimlik bilgisi aktarımı yasağı var mı? | | | |

**B2 = yasak** olan sağlayıcı için abonelik yolu **açılmaz**, teknik olarak
mümkün olsa bile. **B2 = belirsiz** ise kullanıcıya arayüzde uyarı gösterilir.

---

## C. Endpoint ve uyumluluk

| # | Soru | Z.AI | Qwen | Kimi |
| :-- | :--- | :--- | :--- | :--- |
| C1 | OpenAI-uyumlu base URL? | | | |
| C2 | Anthropic-uyumlu endpoint de var mı? | | | |
| C3 | Tool/function calling dokümante ediliyor mu? | | | |
| C4 | Streaming (SSE) destekleniyor mu? | | | |
| C5 | Dokümantasyonda "OpenAI SDK ile kullanın" deniyor mu? | | | |

---

## D. Pratik engeller

| # | Soru | Z.AI | Qwen | Kimi |
| :-- | :--- | :--- | :--- | :--- |
| D1 | Kayıt için Çin telefon numarası / ödeme yöntemi gerekiyor mu? | | | |
| D2 | Bölgesel kısıt var mı? (TR'den erişilebiliyor mu) | | | |
| D3 | Uluslararası ve yerel ayrı platform mu? (ör. farklı domain) | | | |
| D4 | Dokümantasyon İngilizce mevcut mu? | | | |

---

## E. Senin cevaplaman gereken tek karar sorusu

Araştırma bittiğinde şu netleşmeli:

> **Kaç sağlayıcı B tipi çıktı?**

- **0 tanesi** → Abonelik desteği neredeyse bedava. `subscription-auth`
  spec'i büyük ölçüde dokümantasyona iner, P1'e ~0 hafta ekler.
- **1 tanesi** → Tek bir OAuth akışı yazılır. P1'e ~1.5–2 hafta.
- **2–3 tanesi** → Her biri ayrı akış. P1'e ~3 hafta, ve auth adaptör katmanı
  ADR-006'nın "sağlayıcı nötr" iddiasını ciddi biçimde niteler.

Bu sayı, ["iki mod da MVP'de"](../../docs/decisions.md) kararının gerçek
maliyetini belirliyor. Çıkan sayı yüksekse kararı yeniden değerlendirmek
mantıklı olabilir — bu bir geri adım değil, spike'ın işini yapması.

---

## Bulguları nereye yazacağız

- Erişim modeli tablosu → [`docs/protocol.md`](../../docs/protocol.md#acp-capability-matrix)
- ToS pozisyonu + tarihli alıntı → aynı bölüm
- Tip dağılımı → [`acp-orchestrator`](../specs/acp-orchestrator.md) spec'ini
  şekillendirir, sonra ratify edilir

---

## Bu fazın KAPSAMINDA OLMAYAN (anahtar + kod gerekiyor)

Bunları araştırmayla cevaplayamazsın; ölçüm gerekiyor. Sonraki adım:

- Tool şeması: nested object, enum, `additionalProperties: false` kabul ediliyor mu
- Paralel tool call davranışı
- Streaming'de tool delta'larının parçalanma şekli
- Finish reason değerlerinin tam listesi
- Token usage alanlarının streamed yanıtta görünüp görünmediği
- Rate-limit header'ları

Bunlar A/B/C tipi netleştikten sonra, contract test'lerle ölçülecek.
