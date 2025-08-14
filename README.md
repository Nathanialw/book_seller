#!/bin/bash
set -e
echo "Downloading and running ecommerce setup script..."
bash <(curl -sSL https://raw.githubusercontent.com/nathanialw/ecommerce/deploy/)
echo "Setup script finished."

git add .
git commit -m "updated"
git push
git tag v0.0.17
git push origin v0.0.17


go list -m -versions github.com/nathanialw/ecommerce


/run local buffToFind="Ability_ThunderBolt";local found=false;for i=1,16 do local texture=UnitBuff("player",i);if texture and strfind(texture,buffToFind)then found=true;break;end;end if not found then DEFAULT_CHAT_FRAME:AddMessage("Buff not active.");end



/run CastSpellByName("Judgement");
/run local b="Ability_ThunderBolt"; local c="Holy_HolySmite"; local foundBuff; for i=1,16 do local t=UnitBuff("player",i); if t and (strfind(t,b) or strfind(t,c)) then foundBuff=true; break; end; end; if not foundBuff then CastSpellByName("Seal of Righteousness"); end
/cast Crusader Strike
/run if not IsCurrentAction(25) then AttackTarget(); end



/Cast Judgement
/run local b="Ability_ThunderBolt"; local foundBuff; for i=1,16 do local t=UnitBuff("player",i); if t and strfind(t,b) then foundBuff=true; break; end; end; if not foundBuff then CastSpellByName("Seal of Righteousness"); end
/cast Crusader Strike
/run if not IsCurrentAction(25) then AttackTarget(); end






strfind(t,b)or