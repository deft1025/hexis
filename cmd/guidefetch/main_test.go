package main

import (
	"strings"
	"testing"
)

func TestPublicBuildExtractionExcludesStatsAndKeepsCardBoundaries(t *testing.T) {
	body := `<html><title>hero</title><div class="glass-card"><div style="font-size:13px;color:#e2e8f0;margin-bottom:4px">basic</div></div>
<div class="glass-card"><div style="font-size:15px;font-weight:700;color:#fff">build</div><div class="desc-text">description</div><div class="tips-text">tip<br>next</div>
<a href="../augment/core.html">core<span class="pill-win-rate">win rate 99%</span></a><span class="pill-item">item</span><div class="drain-lock-full">locked content</div></div>
<div class="glass-card"><a href="../augment/stat-only.html">stat only</a></div><div class="box-trap"><a href="../augment/trap.html">trap</a> avoid</div></html>`
	p, err := parsePage([]byte(body), "https://lolhaidou.cn/hero/hero.html")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Builds) != 1 || p.Builds[0].Name != "build" || len(p.Builds[0].Augments) != 1 || p.Builds[0].Augments[0].Name != "core" {
		t.Fatalf("incorrect build boundaries: %+v", p.Builds)
	}
	if p.Builds[0].Augments[0].URL != "https://lolhaidou.cn/augment/core.html" || p.Builds[0].Tips != "tip\nnext" {
		t.Fatalf("incorrect link or text: %+v", p.Builds[0])
	}
	if len(p.Basics) != 1 || p.Basics[0] != "basic" {
		t.Fatalf("missing basics: %+v", p.Basics)
	}
	for _, b := range p.Blocks {
		if strings.Contains(b.Text, "99%") || strings.Contains(b.Text, "locked content") {
			t.Fatal("statistics or locked content leaked")
		}
	}
	if p.Blocks[len(p.Blocks)-1].Kind != "box-trap" {
		t.Fatal("negative interaction was not distinguished")
	}
}

func TestEquipmentWithoutDescriptionKeepsRouteAndIgnoresUnrelatedCards(t *testing.T) {
	body := `<div class="glass-card"><div style="font-size:15px;font-weight:700">双发快射暴击特效流</div>
<a href="../augment/doubletap.html">双发快射<span class="pill-win-rate">99%</span></a>
<span class="pill-item">无尽之刃</span><div class="drain-lock-full"><span class="pill-item">隐藏装备</span></div></div>
<div class="glass-card"><div style="font-size:15px;font-weight:700">仅标题</div></div>
<div class="glass-card"><a class="pill-item" href="../hero/lux.html">光辉女郎</a></div>
<div class="glass-card"><a href="../augment/stat.html">统计推荐</a></div>`
	p, err := parsePage([]byte(body), "https://lolhaidou.cn/hero/vayne.html")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Builds) != 0 || len(p.EquipmentBuilds) != 1 {
		t.Fatalf("incorrect equipment boundaries: %+v", p)
	}
	b := p.EquipmentBuilds[0]
	if b.Name != "双发快射暴击特效流" || len(b.Items) != 1 || b.Items[0] != "无尽之刃" || len(b.Augments) != 1 || b.Augments[0].Name != "双发快射" {
		t.Fatalf("lost route or leaked unrelated content: %+v", b)
	}
}
