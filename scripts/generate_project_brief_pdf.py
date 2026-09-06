#!/usr/bin/env python3
"""Cherry proje dosyası PDF — docs/ kaynaklı, spekülasyon yok."""

from __future__ import annotations

from pathlib import Path

from reportlab.lib import colors
from reportlab.lib.enums import TA_CENTER, TA_JUSTIFY, TA_LEFT
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import mm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.platypus import (
    ListFlowable,
    ListItem,
    PageBreak,
    Paragraph,
    SimpleDocTemplate,
    Spacer,
    Table,
    TableStyle,
)

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "cherry-proje-dosyasi.pdf"

INK = colors.HexColor("#0E1114")
PAPER = colors.HexColor("#E8E4DC")
SURFACE = colors.HexColor("#161B20")
BRASS = colors.HexColor("#C4A574")
MUTED = colors.HexColor("#5C656E")
BORDER = colors.HexColor("#2A323A")
DANGER = colors.HexColor("#C45C4A")

FONT_REG = "DejaVu"
FONT_BOLD = "DejaVuBold"
FONT_MONO = "DejaVuMono"

pdfmetrics.registerFont(TTFont(FONT_REG, "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"))
pdfmetrics.registerFont(TTFont(FONT_BOLD, "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"))
pdfmetrics.registerFont(TTFont(FONT_MONO, "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"))


def styles() -> dict[str, ParagraphStyle]:
    base = getSampleStyleSheet()
    s: dict[str, ParagraphStyle] = {}
    s["cover_kicker"] = ParagraphStyle(
        "cover_kicker",
        parent=base["Normal"],
        fontName=FONT_BOLD,
        fontSize=9,
        textColor=BRASS,
        alignment=TA_CENTER,
        letterSpacing=2,
        spaceAfter=8,
    )
    s["cover_title"] = ParagraphStyle(
        "cover_title",
        parent=base["Normal"],
        fontName=FONT_BOLD,
        fontSize=36,
        textColor=PAPER,
        alignment=TA_CENTER,
        leading=42,
        spaceAfter=10,
    )
    s["cover_sub"] = ParagraphStyle(
        "cover_sub",
        parent=base["Normal"],
        fontName=FONT_REG,
        fontSize=12,
        textColor=PAPER,
        alignment=TA_CENTER,
        leading=18,
        spaceAfter=6,
    )
    s["cover_meta"] = ParagraphStyle(
        "cover_meta",
        parent=base["Normal"],
        fontName=FONT_MONO,
        fontSize=9,
        textColor=BRASS,
        alignment=TA_CENTER,
        leading=14,
    )
    s["h1"] = ParagraphStyle(
        "h1",
        parent=base["Normal"],
        fontName=FONT_BOLD,
        fontSize=16,
        textColor=INK,
        spaceBefore=16,
        spaceAfter=8,
        leading=20,
    )
    s["h2"] = ParagraphStyle(
        "h2",
        parent=base["Normal"],
        fontName=FONT_BOLD,
        fontSize=12,
        textColor=INK,
        spaceBefore=12,
        spaceAfter=6,
        leading=16,
    )
    s["body"] = ParagraphStyle(
        "body",
        parent=base["Normal"],
        fontName=FONT_REG,
        fontSize=9.5,
        textColor=INK,
        alignment=TA_JUSTIFY,
        leading=13.5,
        spaceAfter=7,
    )
    s["en"] = ParagraphStyle(
        "en",
        parent=base["Normal"],
        fontName=FONT_REG,
        fontSize=8.5,
        textColor=MUTED,
        alignment=TA_JUSTIFY,
        leading=12,
        spaceAfter=8,
    )
    s["bullet"] = ParagraphStyle(
        "bullet",
        parent=base["Normal"],
        fontName=FONT_REG,
        fontSize=9.5,
        textColor=INK,
        leading=13,
    )
    s["cell"] = ParagraphStyle(
        "cell",
        parent=base["Normal"],
        fontName=FONT_REG,
        fontSize=8,
        textColor=INK,
        leading=11,
    )
    s["cell_h"] = ParagraphStyle(
        "cell_h",
        parent=base["Normal"],
        fontName=FONT_BOLD,
        fontSize=8,
        textColor=PAPER,
        leading=11,
    )
    s["mono"] = ParagraphStyle(
        "mono",
        parent=base["Normal"],
        fontName=FONT_MONO,
        fontSize=8,
        textColor=INK,
        leading=11.5,
        backColor=colors.HexColor("#F3EFE6"),
        borderPadding=6,
        spaceAfter=8,
        leftIndent=4,
        rightIndent=4,
    )
    s["callout"] = ParagraphStyle(
        "callout",
        parent=base["Normal"],
        fontName=FONT_REG,
        fontSize=9,
        textColor=INK,
        leading=13,
        leftIndent=8,
        rightIndent=8,
        spaceBefore=4,
        spaceAfter=8,
    )
    s["footer"] = ParagraphStyle(
        "footer",
        parent=base["Normal"],
        fontName=FONT_MONO,
        fontSize=8,
        textColor=MUTED,
    )
    s["toc"] = ParagraphStyle(
        "toc",
        parent=base["Normal"],
        fontName=FONT_REG,
        fontSize=10,
        textColor=INK,
        leading=16,
        leftIndent=8,
    )
    return s


def P(text: str, style: ParagraphStyle) -> Paragraph:
    return Paragraph(text.replace("\n", "<br/>"), style)


def bullets(items: list[str], st: ParagraphStyle) -> ListFlowable:
    return ListFlowable(
        [ListItem(Paragraph(item, st), leftIndent=12, bulletColor=BRASS) for item in items],
        bulletType="bullet",
        start="•",
        leftIndent=16,
        bulletFontName=FONT_BOLD,
        bulletFontSize=9,
        spaceAfter=8,
    )


def table(headers: list[str], rows: list[list[str]], st: dict[str, ParagraphStyle], col_widths: list[float] | None = None) -> Table:
    head = [Paragraph(h, st["cell_h"]) for h in headers]
    body = [[Paragraph(c, st["cell"]) for c in row] for row in rows]
    t = Table([head, *body], colWidths=col_widths, hAlign="LEFT")
    t.setStyle(
        TableStyle(
            [
                ("BACKGROUND", (0, 0), (-1, 0), SURFACE),
                ("TEXTCOLOR", (0, 0), (-1, 0), PAPER),
                ("BACKGROUND", (0, 1), (-1, -1), colors.white),
                ("FONTNAME", (0, 0), (-1, 0), FONT_BOLD),
                ("VALIGN", (0, 0), (-1, -1), "TOP"),
                ("LEFTPADDING", (0, 0), (-1, -1), 6),
                ("RIGHTPADDING", (0, 0), (-1, -1), 6),
                ("TOPPADDING", (0, 0), (-1, -1), 5),
                ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
                ("GRID", (0, 0), (-1, -1), 0.4, BORDER),
                ("ROWBACKGROUNDS", (0, 1), (-1, -1), [colors.white, colors.HexColor("#F7F3EA")]),
            ]
        )
    )
    return t


def header_footer(canvas, doc) -> None:  # type: ignore[no-untyped-def]
    canvas.saveState()
    w, h = A4
    if doc.page == 1:
        canvas.setFillColor(INK)
        canvas.rect(0, 0, w, h, fill=1, stroke=0)
        canvas.setFillColor(BRASS)
        canvas.rect(0, h - 8, w, 8, fill=1, stroke=0)
        canvas.rect(0, 0, w, 8, fill=1, stroke=0)
        canvas.restoreState()
        return
    canvas.setFillColor(INK)
    canvas.rect(0, h - 14 * mm, w, 14 * mm, fill=1, stroke=0)
    canvas.setFillColor(BRASS)
    canvas.rect(0, h - 14 * mm, w, 1.2, fill=1, stroke=0)
    canvas.setFillColor(PAPER)
    canvas.setFont(FONT_BOLD, 8)
    canvas.drawString(18 * mm, h - 9 * mm, "CHERRY")
    canvas.setFont(FONT_REG, 8)
    canvas.drawRightString(w - 18 * mm, h - 9 * mm, "Proje dosyası  ·  Project dossier")
    canvas.setFillColor(PAPER)
    canvas.rect(0, 0, w, 12 * mm, fill=1, stroke=0)
    canvas.setFillColor(BRASS)
    canvas.rect(0, 12 * mm, w, 0.8, fill=1, stroke=0)
    canvas.setFillColor(MUTED)
    canvas.setFont(FONT_MONO, 8)
    canvas.drawString(18 * mm, 5 * mm, "Kaynak: docs/  ·  spekülasyon yok")
    canvas.drawRightString(w - 18 * mm, 5 * mm, f"{doc.page}")
    canvas.restoreState()


def build() -> None:
    st = styles()
    story: list = []
    usable = A4[0] - 36 * mm

    # Cover content sits on dark page 1
    story.append(Spacer(1, 78 * mm))
    story.append(P("MASAÜSTÜ STÜDYO  ·  DESKTOP STUDIO", st["cover_kicker"]))
    story.append(P("Cherry", st["cover_title"]))
    story.append(
        P(
            "Mobil uygulama yazan atölye — kendisi mobil uygulama değildir.<br/>"
            "A desktop atelier that writes mobile apps — it is not a mobile app.",
            st["cover_sub"],
        )
    )
    story.append(Spacer(1, 18 * mm))
    story.append(
        P(
            "Ne  ·  Nerede  ·  Neden  ·  Nasıl yazıldı<br/>"
            "Nasıl çalıştırılır  ·  Bitti / kalan  ·  Diğer yaklaşımlar",
            st["cover_meta"],
        )
    )
    story.append(Spacer(1, 28 * mm))
    story.append(
        P(
            "Eylül 2026  ·  dilim 1–8 kodda  ·  Windows ve macOS<br/>"
            "Kaynak belgeler: docs/README.md, remaining.md, architecture.md",
            st["cover_meta"],
        )
    )
    story.append(PageBreak())

    story.append(P("1. Kimlik / Identity", st["h1"]))
    story.append(
        P(
            "<b>Cherry</b>, kullanıcının bilgisayarında duran bir masaüstü stüdyodur. "
            "Kişi sohbette brif yazar; arka plandaki ajan seçilen yığında (Expo, Flutter veya SwiftUI) "
            "frontend + backend + Maestro YAML üretir, yerelde ayağa kaldırır, UI test eder. "
            "Müşteriye klasör, zip veya git verilir. Cherry v1’de müşteri uygulamasını barındırmaz.",
            st["body"],
        )
    )
    story.append(
        P(
            "<b>Cherry</b> is a desktop studio on the user’s machine. The person writes a brief in chat; "
            "a background agent produces frontend + backend + Maestro YAML in the chosen stack, boots it locally, "
            "and UI-tests it. Delivery is files (folder / zip / git). Cherry does not host customer apps in v1.",
            st["en"],
        )
    )
    story.append(
        table(
            ["Bu / This", "Değil / Not"],
            [
                ["Windows ve macOS Electron stüdyosu", "Cherry’nin kendisi bir mobil uygulama"],
                ["İki ayrı backend: platform vs müşteri dosyaları", "Tek GraphQL’de her şey karışık"],
                ["Teslim: Expo TS / Flutter Dart / SwiftUI kaynağı", "preview/ HTML sitesi zip’te"],
                ["OpenCode yazıcı; LLM planlar; GDPR sarmalar", "OpenCode’u yeniden yazmak"],
                ["A ve B aynı iş — yoğunluk işçileri", "A = kod, B = test"],
                ["Colab = senin GPU oturumun + dosya", "Cherry içinde 24/7 üretim inferansı"],
                ["X-benzeri güvenlik: şifre, 6 hane, TOTP", "SMS / telefon kimliği"],
            ],
            st,
            [usable * 0.48, usable * 0.52],
        )
    )
    story.append(Spacer(1, 4 * mm))
    story.append(
        P(
            "Marka stüdyo adıdır: <b>Cherry</b>. Eski iç ad (İçerde) kalktı. "
            "UI mürekkep zemin, kâğıt yazı, tek pirinç vurgu — mor “AI chrome” yok.",
            st["body"],
        )
    )

    story.append(P("2. Nerede durur / Where it lives", st["h1"]))
    story.append(
        P(
            "Cherry iki yerde yaşar ve bunlar karıştırılmaz. Platform, stüdyonun kendi hesabı, oturumu, "
            "iş kuyruğu ve denetimidir. Üretilen mobil uygulama diskte ayrı bir ağaçtır.",
            st["body"],
        )
    )
    story.append(
        table(
            ["Parça", "Konum", "Görevi"],
            [
                ["apps/web", "Next.js renderer, port 43147", "Giriş, projeler, stüdyo, LLM, güvenlik, Bağlantılar"],
                ["apps/desktop", "Electron main + preload", "Pencere, cihaz parmak izi, sidecar bin, Maestro MCP"],
                ["services/api", "Go GraphQL, port 43148", "Auth, işler, GDPR, LLM router, yerel aktif çocuk süreç"],
                ["MongoDB", "MONGO_URI yoksa bellek", "Platform durumu — üretilen kaynak burada tutulmaz"],
                ["var/projects/…", "Kullanıcı diski", "frontend/ backend/ maestro/ README — müşteri teslimi"],
                ["vendor/bin", "Kurucu / script", "OpenCode + Maestro + isteğe cloudflared; git’e ikili yok"],
                ["colab/", "Google Colab, senin hesap", "İki notebook (A/B), QLoRA, geçici inferans tüneli"],
            ],
            st,
            [usable * 0.22, usable * 0.30, usable * 0.48],
        )
    )
    story.append(Spacer(1, 3 * mm))
    story.append(
        P(
            "<b>İki backend kuralı.</b> Cherry platformu: Go GraphQL + MongoDB. "
            "Müşteri backend’i: diskteki dosyalar (yerel <font face='DejaVuMono'>go run</font> "
            "127.0.0.1:47000–47999 veya kişinin Supabase / Cloudflare / Render hedefi). "
            "Aynı şemada birleştirilmez. UI asla tarayıcıdan LLM çağırmaz; her çağrı API’den GDPR katmanından geçer.",
            st["body"],
        )
    )
    story.append(
        P(
            "<b>Two backends.</b> Platform = Go GraphQL + MongoDB. Customer backend = files on disk "
            "(local child process on 47000–47999, or the person’s own provider). Never mixed. "
            "The UI never calls an LLM directly.",
            st["en"],
        )
    )

    story.append(P("3. Neden yazıldı / Why it exists", st["h1"]))
    story.append(
        P(
            "Amaç, “sohbet kutusundan HTML site” üretmek değil. Amaç: bir stüdyo sahibinin "
            "kendi makinesinde, seçtiği mobil yığında, Clean Architecture’lı gerçek kaynak almak; "
            "yerelde denemek; Maestro ile UI testi görmek; teslimi dosya olarak vermek. "
            "Barındırma, SMS doğrulama ve müşteri sırrını modele ham göndermek ürünün dışında.",
            st["body"],
        )
    )
    story.append(
        bullets(
            [
                "<b>Kontrol:</b> Kod senin diskine yazılır. Cherry cloud’da müşteri API’si tutmaz.",
                "<b>Yığın seçimi:</b> Expo SDK 57 TypeScript, Flutter 3.47 / Dart 3.13, veya SwiftUI — kişi seçer.",
                "<b>Dürüst test:</b> Cihaz yoksa Maestro SKIPPED. Sahte PASSED yok. OpenCode yoksa iskelet kalır; sahte yazım yok.",
                "<b>KVKK/GDPR:</b> redact → model → tarama → denetim. Eğitim paketi ham e-posta / token taşımaz.",
                "<b>Kapasite:</b> İki LLM yuvası (A ve B) aynı işi yapar; eşzamanlı yük içindir, rol ayrımı değil.",
            ],
            st["bullet"],
        )
    )
    story.append(
        P(
            "Why: real source in a chosen mobile stack, local activate, honest Maestro results, file handoff — "
            "not hosted customer backends, not SMS identity, not unredacted PII to a model.",
            st["en"],
        )
    )

    story.append(P("4. Nasıl yazıldı / How it was built", st["h1"]))
    story.append(
        P(
            "Karar: önce çalışan kabuk, sonra tek LLM yuvası. İki model, Colab ve fine-tune en başa alınmadı. "
            "Yapım sırası belgede dilim 1→8. Dilim 1–8 kodda bitti.",
            st["body"],
        )
    )
    story.append(
        table(
            ["Dilim", "Ne yapıldı"],
            [
                ["1 İskele", "Electron + Next.js (tasarım token) + Go GraphQL + Mongo. Boş proje hali."],
                ["2 Auth + posta", "Şifre, 6 hane, birinci parti mailer, güvenilir cihaz, tek oturum, TOTP. SMS yok."],
                ["3 Proje diski", "Brif, yığın, frontend/ backend/ maestro/, zip. preview/ teslim değil."],
                ["4 LLM A + GDPR", "Tek işçi. redact → tamamla → tarama. Admin’de aktif versiyon pointer."],
                ["5 OpenCode", "opencode run --dir (mutlak yol) + --auto. TUI yok. CLI yoksa iskelet."],
                ["6 Yerel + Maestro", "Çocuk API 47000–47999. Sidecar vendor/bin. Cihaz yok → SKIPPED."],
                ["7 LLM B + kuyruk", "A ile aynı iş. Boş yuva alır; ikisi meşgulse kuyruk. Pointer sonraki işi değiştirir."],
                ["8 Colab fine-tune", "İki notebook, seed paket, eğitim zip → registerLlmVersion. Üretim API değil."],
            ],
            st,
            [usable * 0.22, usable * 0.78],
        )
    )
    story.append(Spacer(1, 3 * mm))
    story.append(
        P(
            "Kod sözleşmesi: her bölümün <font face='DejaVuMono'>.cursor/rules</font> kuralı ve "
            "<font face='DejaVuMono'>docs/</font> belgesi vardır. Belge yoksa o bölüm yazılmaz. "
            "Davranış değişince belge aynı değişiklikte güncellenir. TypeScript birleşiklerde "
            "<font face='DejaVuMono'>default</font> + <font face='DejaVuMono'>never</font>. "
            "Lorem ipsum yok; boş / yükleniyor / hata her ekranda.",
            st["body"],
        )
    )

    story.append(P("5. Teknoloji yığını / Stack", st["h1"]))
    story.append(
        table(
            ["Katman", "Seçim", "Not"],
            [
                ["Stüdyo UI", "Next.js 16 + TS + Tailwind + shadcn", "Electron renderer; Cherry mobil UI yok"],
                ["Masaüstü", "Electron (Win/Mac)", "nodeIntegration kapalı; dar preload"],
                ["Platform API", "Go + gqlgen GraphQL", "Tek sözleşme; health dışında ad-hoc REST yok"],
                ["Platform DB", "MongoDB veya bellek", "MONGO_URI boşsa restart’ta oturum gider"],
                ["Yazıcı", "OpenCode CLI (vendor)", "Yeniden yazılmaz; cherry-colab provider Colab /v1/chat/completions"],
                ["UI test", "Maestro CLI + MCP", "Java 17+; emülatör yoksa SKIPPED"],
                ["Müşteri FE", "Expo 57 / Flutter 3.47 / SwiftUI", "Clean Architecture katmanları zorunlu"],
                ["Müşteri BE", "Yerel Go veya Bağlantılar hedefi", "Kişinin Supabase/Cloudflare/Render hesabı"],
                ["Posta", "SMTP veya Resend", "Yoksa in-app kutu; AgentMail yok"],
                ["Colab", "Qwen 1.5B QLoRA, T4 16GB", "İki oturum = iki kart; named Cloudflare tüneli isteğe"],
            ],
            st,
            [usable * 0.20, usable * 0.32, usable * 0.48],
        )
    )

    story.append(P("6. Nasıl çalıştırılır / How to run", st["h1"]))
    story.append(
        P(
            "Depo npm workspaces monorepo’dur. Kurulum yalnızca kökte. "
            "<font face='DejaVuMono'>apps/*</font> içinde <font face='DejaVuMono'>npm install</font> yok.",
            st["body"],
        )
    )
    story.append(
        P(
            "npm install<br/>"
            "npm run dev:api&nbsp;&nbsp;&nbsp;# Go GraphQL → 127.0.0.1:43148<br/>"
            "npm run dev:web&nbsp;&nbsp;&nbsp;# Next.js → 127.0.0.1:43147<br/>"
            "npm run dev:desktop&nbsp;&nbsp;# Electron kabuk (isteğe)",
            st["mono"],
        )
    )
    story.append(
        bullets(
            [
                "UI: http://127.0.0.1:43147 — hesap oluştur (şifre ≥ 8). Yeni cihaz: 6 haneli kod (in-app kutu veya SMTP).",
                "OpenCode + Maestro içeri: <font face='DejaVuMono'>./scripts/vendor-sidecars.sh</font> — resmi CLI iner, git’e ikili konmaz. API’yi yeniden başlat.",
                "Colab inferans: notebook hücre 8–9 (FastAPI :8000 + cloudflared). LLM yönetici → işçi A/B URL (<font face='DejaVuMono'>…/v1</font>) → Kaydet. Token yalnızca Colab secret.",
                "Gerçek posta: <font face='DejaVuMono'>.env</font> içinde SMTP_* veya RESEND_API_KEY. Kalıcılık: MONGO_URI.",
                "OAuth (isteğe): GitHub/Vercel/Supabase/Render client id. Yoksa yerel izin ekranı — sessiz “bağlandı” yok.",
            ],
            st["bullet"],
        )
    )
    story.append(
        P(
            "Arama sırası sidecar: CHERRY_OPENCODE_BIN → CHERRY_SIDECAR_DIR (Electron resources/bin) → vendor/bin → PATH. "
            "CLI yetmez: model için bağlı Colab URL veya CHERRY_LLM_API_KEY. İkisi de yoksa yazım düşer; sahte dosya yok.",
            st["body"],
        )
    )

    story.append(P("7. Güvenlik ve KVKK / Security and GDPR", st["h1"]))
    story.append(
        P(
            "Akış X’e benzer: e-posta + şifre; yeni veya güvenilmeyen cihazda 6 hane; isteğe TOTP. "
            "Tek aktif oturum — yeni giriş diğerlerini düşürür. Telefon numarası kimlik değildir. SMS yoktur.",
            st["body"],
        )
    )
    story.append(
        P(
            "Her LLM çağrısı: girdi redaksiyonu (e-posta, telefon, TCKN, kod, token, müşteri .env sırları) → "
            "işçi A veya B → çıktı taraması → denetim kaydı. "
            "<font face='DejaVuMono'>exportMe</font> / <font face='DejaVuMono'>deleteMe</font> birinci sınıf. "
            "Colab eğitim paketi <font face='DejaVuMono'>exportMe</font> değildir; yalnızca redakte SFT satırları.",
            st["body"],
        )
    )
    story.append(
        P(
            "Security is X-inspired: password, 6-digit new-device code, TOTP, trusted devices, one active session. "
            "No SMS. Every model call is redact → complete → scan → audit.",
            st["en"],
        )
    )

    story.append(P("8. Ajan hattı / Agent path", st["h1"]))
    story.append(
        P(
            "Kişi yalnızca Cherry sohbetine yazar. OpenCode TUI açılmaz. "
            "GraphQL <font face='DejaVuMono'>createProject</font> / <font face='DejaVuMono'>sendProjectMessage</font> → "
            "GDPR’li plan (LLM) → <font face='DejaVuMono'>opencode run --dir --auto</font> "
            "(sonraki mesajda --continue) → isteğe yerel API + Maestro.",
            st["body"],
        )
    )
    story.append(
        table(
            ["Adım", "Ne olur", "Dürüst başarısızlık"],
            [
                ["Brif + yığın", "İskelet Clean Architecture ağacı", "Kısa brif → doğrulama hatası"],
                ["LLM plan", "llm/plan.md, kanal mock / http / colab-tunnel", "Colab kapalıysa DISCONNECTED, varsayılan completer"],
                ["OpenCode", "Seçilen dilde kaynak genişletme", "CLI yok → iskelet; anahtar yok → hata, sahte yazım yok"],
                ["Yerel aktif", "Çocuk süreç localhost 47xxx", "Public URL yok"],
                ["Maestro", "YAML akış, cihaz varsa koşum", "CLI veya emülatör yok → SKIPPED, PASSED değil"],
                ["Zip / git", "frontend+backend+maestro, dil kaynağı", "preview/ HTML zip’e girmez"],
            ],
            st,
            [usable * 0.20, usable * 0.42, usable * 0.38],
        )
    )
    story.append(Spacer(1, 3 * mm))
    story.append(
        P(
            "Colab inferans geçicidir: oturum kapanınca tünel 530 verir. Named tunnel token’ı Cherry API sürecine konmaz. "
            "Paket/adapter köprüsü ayrı bir loopback + quick tunnel’dır (GraphQL’i internete açmaz).",
            st["body"],
        )
    )

    story.append(P("9. Teslim / Handoff", st["h1"]))
    story.append(
        table(
            ["Yığın", "Dil", "Mimari örnek"],
            [
                ["Expo", "SDK 57, TypeScript strict, RN 0.86", "frontend/src/{domain,data,presentation,app}"],
                ["Flutter", "3.47 / Dart 3.13", "lib/features/&lt;özellik&gt;/{domain,data,presentation}"],
                ["SwiftUI (NATIVE)", "Swift 6, iOS 18+", "Domain / Data / Presentation / App"],
            ],
            st,
            [usable * 0.22, usable * 0.38, usable * 0.40],
        )
    )
    story.append(Spacer(1, 3 * mm))
    story.append(
        P(
            "Backend hedefi Bağlantılar’dan: LOCAL (varsayılan), SUPABASE, CLOUDFLARE, RENDER. "
            "GitHub push kişinin reposunadır. Cherry host değildir.",
            st["body"],
        )
    )

    story.append(P("10. Bitti ve kalan / Done and remaining", st["h1"]))
    story.append(
        P(
            "Takvim tahmini (ay/yıl) bu dosyada yok — o spekülasyondur. Durum teknik dilimlere göredir. "
            "Dilim 1–8 kodda. Colab üretim inferansı değildir.",
            st["body"],
        )
    )
    story.append(
        P(
            "No calendar-year roadmap here. Status is by technical slice. Slices 1–8 are in code. "
            "Colab is not production inference.",
            st["en"],
        )
    )
    story.append(P("Kodda olan / In code", st["h2"]))
    story.append(
        bullets(
            [
                "Kabuk, auth (6 hane / TOTP / tek oturum), proje diski, zip kaynak dilde.",
                "GDPR sarmalı, LLM A/B kuyruk, versiyon pointer, OpenCode sidecar, yerel aktif, Maestro SKIPPED/PASSED yolu.",
                "Mongo adaptörü (URI varsa), OAuth sağlayıcı kablosu, Colab A/B ayrı inferans URL, seed paket notebook’ta gömülü.",
                "Named Cloudflare tüneli: token Colab secret; stüdyo yalnızca public HTTPS saklar.",
            ],
            st["bullet"],
        )
    )
    story.append(P("Bilerek yapılmayan / Intentionally not in v1", st["h2"]))
    story.append(
        bullets(
            [
                "Müşteri backend’ini Cherry’nin host etmesi.",
                "Cherry’nin kendisi için mobil istemci.",
                "SMS / telefon kimliği / AgentMail / Figma MCP.",
                "Colab’i stüdyo içinde 24/7 üretim API yapmak.",
                "Tek T4 kartta A+B notebook.",
            ],
            st["bullet"],
        )
    )
    story.append(P("Senin makinede doğrulanacak / Verify on your machine", st["h2"]))
    story.append(
        bullets(
            [
                "MONGO_URI ile kalıcılık; yoksa bellek.",
                "Emülatör açıkken Maestro PASSED (yoksa SKIPPED doğru).",
                "OpenCode’un seçilen yığında gerçek dosya yazması — küçük Colab 1.5B plan üretebilir, büyük uygulama için yetersiz kalabilir; bu kablo değil model kapasitesi.",
                "İşçi B için ikinci named tunnel / ikinci URL.",
                "Electron installer: vendor/bin → resources/bin, Win/Mac paket.",
                "Organizasyon ekranı: sidebar’da şu an kapalı (enabled: false).",
                "CI / GitHub Release henüz yok.",
            ],
            st["bullet"],
        )
    )

    story.append(P("11. Diğer yaklaşımlarla fark / Versus other approaches", st["h1"]))
    story.append(
        table(
            ["Yaklaşım", "Cherry farkı"],
            [
                [
                    "v0 / Lovable / “HTML site sohbeti”",
                    "Teslim preview HTML değil; Expo/Flutter/SwiftUI + Clean Architecture. Zip’te preview/ yok.",
                ],
                [
                    "FlutterFlow / no-code builder",
                    "Kaynak ajanla yazılır, dışa kilitli canvas değil. Kişi yığını seçer.",
                ],
                [
                    "Firebase / Vercel’de “biz host ederiz”",
                    "v1 host yok. Bağlantılar = kişinin hesabı. Yerel aktif yalnızca localhost.",
                ],
                [
                    "Cursor’u ürün sanmak",
                    "Cursor Cherry’yi yazmak içindir. Ürün Electron stüdyosu + Go API’dir. Müşteri Cursor kurmaz.",
                ],
                [
                    "OpenCode’u forklamak",
                    "Cherry CLI çağırır. TUI gizlenir. Eksik CLI = iskelet, uydurma yazım yok.",
                ],
                [
                    "Tek LLM kutusu",
                    "İki kapasite yuvası + kuyruk + immutable versiyon + GDPR zorunlu sarmalayıcı.",
                ],
                [
                    "Colab’i “sunucu” sanmak",
                    "Fine-tune ve geçici tünel. Oturum bitince kopar. Platform işçisi yerelde / HTTP anahtardadır.",
                ],
                [
                    "SMS OTP ürünleri",
                    "Kimlik e-posta + şifre + 6 hane + TOTP. Telefon yok.",
                ],
            ],
            st,
            [usable * 0.32, usable * 0.68],
        )
    )

    story.append(P("12. Stüdyoda doğrulanan durum / Lab check", st["h1"]))
    story.append(
        P(
            "Eylül 2026’da bu depoda uçtan uca bakıldı (bellek store, in-app kutu). "
            "Bu bir üretim sertifikası değil; kablonun kapalı olmadığını gösterir.",
            st["body"],
        )
    )
    story.append(
        table(
            ["Kontrol", "Gözlenen"],
            [
                ["API + UI", "43148 / 43147 ayakta"],
                ["Auth", "Kayıt → 6 hane → oturum"],
                ["Colab named URL", "GET /v1/models 200; POST /v1/chat/completions 200 (oturum açıkken)"],
                ["setColabInferenceUrl A", "CONNECTED — kanal colab-tunnel"],
                ["Proje boru hattı", "GDPR → işçi A (colab-tunnel) → OpenCode yazdı (vendor/bin, v1.18.29) → READY"],
                ["Maestro", "CLI bundled; emülatör yok → SKIPPED"],
                ["1.5B çıktı niteliği", "CLI koştu; küçük model Expo kaynaklarını anlamlı genişletmedi — kapasite notu"],
            ],
            st,
            [usable * 0.32, usable * 0.68],
        )
    )

    story.append(P("13. Okuma haritası / Read map", st["h1"]))
    story.append(
        bullets(
            [
                "docs/README.md — dizin",
                "docs/remaining.md — bitti / kalan",
                "docs/architecture.md — iki backend, süreç sınırları",
                "docs/build-order.md — dilimler",
                "docs/security.md + email-verification.md",
                "docs/mobile-factory.md + opencode.md + maestro.md + local-activate.md",
                "docs/llmops.md + colab.md + gdpr-kvkk.md",
                "docs/connections.md — Bağlantılar / OAuth",
                "docs/design-system.md + screens.md + motion.md",
                "AGENTS.md — ajanlar için sabitler",
                "Bu PDF belgelerden üretilir. Çelişirse docs/ ve kod. Yenile: python3 scripts/generate_project_brief_pdf.py",
            ],
            st["bullet"],
        )
    )

    OUT.parent.mkdir(parents=True, exist_ok=True)
    doc = SimpleDocTemplate(
        str(OUT),
        pagesize=A4,
        leftMargin=18 * mm,
        rightMargin=18 * mm,
        topMargin=20 * mm,
        bottomMargin=16 * mm,
        title="Cherry — Proje dosyası / Project dossier",
        author="Cherry",
        subject="Ne, nerede, neden, nasıl; çalıştırma; bitti/kalan; diğer yaklaşımlar",
    )
    doc.build(story, onFirstPage=header_footer, onLaterPages=header_footer)
    print(f"wrote {OUT} ({OUT.stat().st_size} bytes)")


if __name__ == "__main__":
    build()
