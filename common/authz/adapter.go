// admin/common/authz/adapter.go
package authz

import (
	"context"
	"errors"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/yc-alpha/admin/ent"
	"github.com/yc-alpha/admin/ent/casbinrule"
)

type Adapter struct {
	client *ent.Client
}

func NewAdapter(client *ent.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) LoadPolicy(model model.Model) error {
	ctx := context.Background()
	rules, err := a.client.CasbinRule.Query().All(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		loadPolicyLine(rule, model)
	}
	return nil
}

func (a *Adapter) SavePolicy(model model.Model) error {
	ctx := context.Background()

	// 清空现有策略
	_, err := a.client.CasbinRule.Delete().Exec(ctx)
	if err != nil {
		return err
	}

	var lines []string

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			lines = append(lines, ptype+", "+strings.Join(rule, ", "))
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			lines = append(lines, ptype+", "+strings.Join(rule, ", "))
		}
	}

	return a.saveLines(ctx, lines)
}

func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()
	return a.saveLine(ctx, ptype, rule)
}

func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()
	query := a.client.CasbinRule.Delete().Where(casbinrule.PtypeEQ(ptype))

	for i, v := range rule {
		switch i {
		case 0:
			query = query.Where(casbinrule.V0EQ(v))
		case 1:
			query = query.Where(casbinrule.V1EQ(v))
		case 2:
			query = query.Where(casbinrule.V2EQ(v))
		case 3:
			query = query.Where(casbinrule.V3EQ(v))
		case 4:
			query = query.Where(casbinrule.V4EQ(v))
		case 5:
			query = query.Where(casbinrule.V5EQ(v))
		}
	}

	_, err := query.Exec(ctx)
	return err
}

func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}

func (a *Adapter) saveLine(ctx context.Context, ptype string, rule []string) error {
	create := a.client.CasbinRule.Create().SetPtype(ptype)

	if len(rule) > 0 {
		create = create.SetV0(rule[0])
	}
	if len(rule) > 1 {
		create = create.SetV1(rule[1])
	}
	if len(rule) > 2 {
		create = create.SetV2(rule[2])
	}
	if len(rule) > 3 {
		create = create.SetV3(rule[3])
	}
	if len(rule) > 4 {
		create = create.SetV4(rule[4])
	}
	if len(rule) > 5 {
		create = create.SetV5(rule[5])
	}

	_, err := create.Save(ctx)
	return err
}

func (a *Adapter) saveLines(ctx context.Context, lines []string) error {
	for _, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) < 2 {
			continue
		}
		ptype := parts[0]
		rule := parts[1:]
		if err := a.saveLine(ctx, ptype, rule); err != nil {
			return err
		}
	}
	return nil
}

func loadPolicyLine(rule *ent.CasbinRule, model model.Model) {
	lineText := rule.Ptype
	if rule.V0 != "" {
		lineText += ", " + rule.V0
	}
	if rule.V1 != "" {
		lineText += ", " + rule.V1
	}
	if rule.V2 != "" {
		lineText += ", " + rule.V2
	}
	if rule.V3 != "" {
		lineText += ", " + rule.V3
	}
	if rule.V4 != "" {
		lineText += ", " + rule.V4
	}
	if rule.V5 != "" {
		lineText += ", " + rule.V5
	}

	persist.LoadPolicyLine(lineText, model)
}
